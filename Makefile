shed: static-files
	go build -o ./cmd/web/shed cmd/web/main.go

static-files:
	-[ ! -L "$(pwd)/server/build" ] && ln -sv "../cmd/web/build" ./server
	-[ ! -L "$(pwd)/server/static" ] && ln -sv "../cmd/web/static" ./server
