FROM golang:1.19

WORKDIR /qms/

COPY . /qms/

RUN go mod download
RUN go build -o qms ./cmd/qms/
RUN ls

CMD [ "./qms" ]