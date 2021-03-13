FROM golang:1.15

LABEL author="Mina Gyimah"

WORKDIR /app

ADD . /app

RUN go mod download

RUN go build -o ./cmd/web/shed cmd/web/main.go

WORKDIR /app/cmd/web

RUN go install

EXPOSE 8000

ENTRYPOINT ["/app/cmd/web/shed"]
