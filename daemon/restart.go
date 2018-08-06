package daemon

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/container"
	"github.com/docker/docker/libcontainerd"
	"github.com/docker/docker/runconfig"
	"google.golang.org/grpc"
	"strings"
	"syscall"
)

// ContainerRestart stops and starts a container. It attempts to
// gracefully stop the container within the given timeout, forcefully
// stopping it if the timeout is exceeded. If given a negative
// timeout, ContainerRestart will wait forever until a graceful
// stop. Returns an error if the container cannot be found, or if
// there is an underlying error at any stage of the restart.
func (daemon *Daemon) ContainerRestart(name string, seconds int) error {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}
	if err := daemon.containerRestart(container, seconds); err != nil {
		return fmt.Errorf("Cannot restart container %s: %v", name, err)
	}
	return nil
}

// containerRestart attempts to gracefully stop and then start the
// container. When stopping, wait for the given duration in seconds to
// gracefully stop, before forcefully terminating the container. If
// given a negative duration, wait forever for a graceful stop.
func (daemon *Daemon) containerRestart(container *container.Container, seconds int) error {
	// Avoid unnecessarily unmounting and then directly mounting
	// the container when the container stops and then starts
	// again
	if err := daemon.Mount(container); err == nil {
		defer daemon.Unmount(container)
	}

	if err := daemon.containerStop(container, seconds); err != nil {
		return err
	}

	//container.RestartCount = container.RestartCount + 1
	if err := daemon.restart(container); err != nil {
		return err
	}

	daemon.LogContainerEvent(container, "restart")
	return nil
}

func (daemon *Daemon) restart(container *container.Container) (err error) {
	container.Lock()
	defer container.Unlock()

	restartCount := container.RestartCount + 1
	if container.Running {
		return nil
	}

	if container.RemovalInProgress || container.Dead {
		return fmt.Errorf("Container is marked for removal and cannot be started.")
	}

	// if we encounter an error during start we need to ensure that any other
	// setup has been cleaned up properly
	defer func() {
		if err != nil {
			container.SetError(err)
			// if no one else has set it, make sure we don't leave it at zero
			if container.ExitCode() == 0 {
				container.SetExitCode(128)
			}
			container.ToDisk()
			daemon.Cleanup(container)
		}
	}()

	if err := daemon.conditionalMountOnStart(container); err != nil {
		return err
	}

	// Make sure NetworkMode has an acceptable value. We do this to ensure
	// backwards API compatibility.
	container.HostConfig = runconfig.SetDefaultNetModeIfBlank(container.HostConfig)

	if err := daemon.initializeNetworking(container); err != nil {
		return err
	}

	spec, err := daemon.createSpec(container)
	if err != nil {
		return err
	}

	createOptions := []libcontainerd.CreateOption{libcontainerd.WithRestartManager(container.RestartManager(true))}
	copts, err := daemon.getLibcontainerdCreateOptions(container)
	if err != nil {
		return err
	}
	if copts != nil {
		createOptions = append(createOptions, *copts...)
	}

	if err := daemon.containerd.Create(container.ID, *spec, createOptions...); err != nil {
		errDesc := grpc.ErrorDesc(err)
		logrus.Errorf("Create container failed with error: %s", errDesc)
		// if we receive an internal error from the initial start of a container then lets
		// return it instead of entering the restart loop
		// set to 127 for container cmd not found/does not exist)
		if strings.Contains(errDesc, container.Path) &&
			(strings.Contains(errDesc, "executable file not found") ||
				strings.Contains(errDesc, "no such file or directory") ||
				strings.Contains(errDesc, "system cannot find the file specified")) {
			container.SetExitCode(127)
		}
		// set to 126 for container cmd can't be invoked errors
		if strings.Contains(errDesc, syscall.EACCES.Error()) {
			container.SetExitCode(126)
		}

		container.Reset(false)

		return fmt.Errorf("%s", errDesc)
	}

	container.RestartCount = restartCount
	return container.ToDisk()

}
