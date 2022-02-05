##
## BUILD
## 
FROM golang:1.17-bullseye AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY *.go ./

RUN go build -o /stomper

##
## Deploy
##
FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=build /stomper /app/stomper

EXPOSE 32801

USER nonroot:nonroot

ENTRYPOINT [ "/app/stomper" ]
