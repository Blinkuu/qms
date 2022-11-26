FROM alpine:3

ARG TARGETARCH

WORKDIR /app/

ADD bin/sut bin/sut

ENTRYPOINT [ "./bin/sut" ]