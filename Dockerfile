FROM golang:bookworm as build
ARG VERSION
ENV VLD_VERSION=${VERSION:-"development"}
RUN go install github.com/go-task/task/v3/cmd/task@latest
RUN mkdir /vld
ADD . /vld
WORKDIR /vld
RUN task build

FROM debian:bookworm-slim as publish
ARG VERSION
ENV VLD_VERSION=${VERSION:-"development"}
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install --no-install-recommends -y ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
COPY --from=build /vld/bin /vld
RUN ln -s /vld/vld /usr/local/bin/vld
WORKDIR /vld
CMD ["vld"]