FROM golang:1.18 AS builder

# build paperless-launcher
WORKDIR /build
COPY . .
ENV GO111MODULE="on" \
  GOARCH="amd64" \
  GOOS="linux" \
  CGO_ENABLED="0"
RUN go build


FROM debian:bullseye-slim

# docker-in-docker
ENV DOCKER_EXTRA_OPTS '--storage-driver=overlay2'
COPY entrypoint.sh /usr/local/bin/

RUN apt-get update && apt-get install -y --no-install-recommends \
  apt-transport-https \
  ca-certificates \
  curl \
  gnupg2 \
  software-properties-common && \
  curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - && \
  add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/debian \
   $(lsb_release -cs) stable" && \
  apt-get update && apt-get install -y --no-install-recommends docker-ce && \
  docker -v && \
  dockerd -v && \
  curl -fL -o /usr/local/bin/dind "https://raw.githubusercontent.com/moby/moby/52379fa76dee07ca038624d639d9e14f4fb719ff/hack/dind" && \
	chmod +x /usr/local/bin/dind && \
  chmod +x /usr/local/bin/entrypoint.sh

VOLUME /var/lib/docker
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# paperless-launcher
WORKDIR /app
RUN apt-get install -y wget sudo libfuse2 && \
  wget https://launchpad.net/veracrypt/trunk/1.25.9/+download/veracrypt-console-1.25.9-Debian-11-amd64.deb && \
  dpkg -i veracrypt-console-1.25.9-Debian-11-amd64.deb && rm veracrypt-console-1.25.9-Debian-11-amd64.deb && \
  mkdir /app/mount

ENV PLL_SERVE=0.0.0.0:3000 \
  PLL_MOUNT_PATH=/app/mount/%user% \
  PLL_VOLUME_PATH=/data/%user%.hc

COPY --from=builder /build/paperless-launcher /app
RUN chmod +x /app/paperless-launcher

VOLUME /data
CMD ["/app/paperless-launcher"]