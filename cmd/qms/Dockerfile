FROM alpine:3

ARG TARGETARCH

WORKDIR /app/

ADD bin/qms bin/qms

ENTRYPOINT [ "./bin/qms" ]