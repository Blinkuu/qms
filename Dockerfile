FROM golang:1.19

WORKDIR /qms/

COPY . /qms/

RUN go mod download
RUN go build -o httpserver ./cmd/httpserver/
RUN ls

CMD [ "./httpserver" ]