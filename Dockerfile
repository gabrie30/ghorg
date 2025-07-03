# ==================================================
# Build image
# ==================================================
FROM golang:alpine AS build-image

# ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/gabrie30/ghorg

COPY . .

RUN go get -d -v ./... \
    && go build -a --mod vendor -o ghorg .

# ==================================================
# Runtime image
# ==================================================
FROM alpine:latest AS runtime-image

ARG USER=ghorg
ARG GROUP=ghorg
ARG UID=1111
ARG GID=2222

ENV XDG_CONFIG_HOME=/config
ENV GHORG_CONFIG=/config/conf.yaml
ENV GHORG_RECLONE_PATH=/config/reclone.yaml
ENV GHORG_ABSOLUTE_PATH_TO_CLONE_TO=/data

RUN apk add -U --no-cache ca-certificates openssh-client tzdata git curl tini \
    && mkdir -p /data $XDG_CONFIG_HOME \
    && addgroup --gid $GID $GROUP \
    && adduser -D -H --gecos "" \
                     --home "/home" \
                     --ingroup "$GROUP" \
                     --uid "$UID" \
                     "$USER" \
    && chown -R $USER:$GROUP /home /data $XDG_CONFIG_HOME \
    && rm -rf /tmp/* /var/{cache,log}/* /var/lib/apt/lists/*

USER $USER
WORKDIR /data

# Sample config
COPY --from=build-image --chown=$USER:$GROUP /go/src/github.com/gabrie30/ghorg/sample-conf.yaml /config/conf.yaml
COPY --from=build-image --chown=$USER:$GROUP /go/src/github.com/gabrie30/ghorg/sample-reclone.yaml /config/reclone.yaml

# Copy compiled binary
COPY --from=build-image --chown=$USER:$GROUP /go/src/github.com/gabrie30/ghorg/ghorg /usr/local/bin

VOLUME /data

ENTRYPOINT ["/sbin/tini", "--", "ghorg"]
CMD ["--help"]
