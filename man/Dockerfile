FROM    alpine:3.4
RUN     sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN     apk add -U git go bash curl gcc musl-dev make

RUN     mkdir -p /go/src /go/bin /go/pkg
ENV     GOPATH=/go
RUN     export GLIDE=v0.11.1; \
        export TARGET=/go/src/github.com/Masterminds; \
        mkdir -p ${TARGET} && \
        git clone https://github.com/Masterminds/glide.git ${TARGET}/glide && \
        cd ${TARGET}/glide && \
        make build && \
        cp ./glide /usr/bin/glide && \
        cd / && rm -rf /go/src/* /go/bin/* /go/pkg/*

COPY    glide.yaml /manvendor/
COPY    glide.lock /manvendor/
COPY    glide.yaml /root/.glide/
COPY    glide.lock /root/.glide/
WORKDIR /manvendor/
RUN     glide mirror set https://golang.org/x/sys https://github.com/golang/sys --vcs git
RUN     glide mirror set https://golang.org/x/image https://github.com/golang/image --vcs git
RUN     glide install && mv vendor src
ENV     GOPATH=$GOPATH:/go/src/github.com/docker/docker/vendor:/manvendor
RUN     go build -o /usr/bin/go-md2man github.com/cpuguy83/go-md2man

WORKDIR /go/src/github.com/docker/docker/
ENTRYPOINT ["man/generate.sh"]
