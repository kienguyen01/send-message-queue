FROM golang:1.19.3

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./


RUN go build -o /go-output

EXPOSE 8080

CMD ['/go-output']

