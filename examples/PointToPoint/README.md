## Point-to-point Messaging

This is a simple example that uses a stomper instance to pass messages from one publisher to one subscriber.

To run this example, navigate to this folder on your local machine and run `docker-compose up`. This will build and run 3 containers. The publisher and subscriber will loop until the server comes up, then connect. Then the publisher will send messages every 10 seconds. The subscriber will record them to a file.

You can observe this process in two ways: if you run just `docker-compose up`, the stomper server will write its logs to STDOUT and you can see the operation there. If you run `docker-compose up -d` to detach, then you can use `docker exec` to spawn a shell instance in the subscriber container and observe the messages written to the file there.

N.b. both the stomper image and the publisher image are built using a multi-stage process based on Google's distroless images to minimize size. The subscriber image is built on the full `golang:1.17-alpine` image to allow easy access via `docker exec`.

To test with larger numbers of clients, can run a command like:

```shell
$ docker-compose up  --scale publisher=5 --scale subscriber=3
```
