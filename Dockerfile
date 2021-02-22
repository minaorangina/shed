FROM golang:1.15

LABEL author="Mina Gyimah"

WORKDIR /app

ADD . /app

RUN go mod download

RUN make

EXPOSE 8000

ENTRYPOINT ["/app/cmd/web/shed"]
