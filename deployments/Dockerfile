FROM golang:1.19

WORKDIR /qms/

ADD . .

RUN go build -i -o qms ./cmd/qms

CMD [ "./qms" ]