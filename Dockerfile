FROM golang:alpine

# Install dependencies for copy
RUN apk add -U --no-cache ca-certificates tzdata git

WORKDIR /go/src/github.com/gabrie30/ghorg
COPY . .

# Fetching dependencies and build the app
RUN go get -d -v ./... \
    && CGO_ENABLED=0 go build -a --mod vendor -o ghorg .

# Needed for reclone command
RUN cp ./ghorg /go/bin/ghorg
