##
## BUILD
## 
FROM golang:1.17-bullseye AS build

WORKDIR /app

COPY go.mod ./

RUN go mod download
COPY *.go ./

RUN go build -o /client

##
## Deploy
##
FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=build /client /app/client

EXPOSE 32801

ENTRYPOINT [ "/app/client" ]
