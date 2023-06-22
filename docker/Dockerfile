# Dockerfile References: https://docs.docker.com/engine/reference/builder/

FROM golang:1.20-alpine

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# Add Maintainer Info
LABEL maintainer="Chernigin Alexey <thenorth@bk.ru>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum  /

RUN pwd
RUN echo "9999999"

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
#COPY ./ .


# Expose port 8080 to the outside world
EXPOSE 1234



RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon




