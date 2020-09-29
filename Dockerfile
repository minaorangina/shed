FROM golang:1.15

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go mod download

RUN cd cmd/web && GOOS=linux GOARCH=amd64 go build -o shed .

CMD ["/app/cmd/web/shed"]
