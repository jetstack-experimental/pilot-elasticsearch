FROM alpine:3.6

ADD pilot_linux_amd64 /pilot

ENTRYPOINT ["/pilot"]
