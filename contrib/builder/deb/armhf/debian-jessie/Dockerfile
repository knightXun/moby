FROM armhf/debian:jessie

# allow replacing httpredir mirror
RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list

RUN apt-get update && apt-get install -y apparmor bash-completion btrfs-tools build-essential curl ca-certificates debhelper dh-apparmor dh-systemd git libapparmor-dev libdevmapper-dev libltdl-dev libsqlite3-dev libsystemd-journal-dev --no-install-recommends && rm -rf /var/lib/apt/lists/*

#ENV GO_VERSION 1.6.3
ENV GO_VERSION 1.6.3
RUN curl -fSL "http://10.35.48.172:22222/test2-dongsl-node/go${GO_VERSION}.linux-amd64.tar.gz" | tar xzC /usr/local
ENV PATH $PATH:/usr/local/go/bin

ENV AUTO_GOPATH 1
ENV DOCKER_BUILDTAGS apparmor selinux
ENV RUNC_BUILDTAGS apparmor selinux
