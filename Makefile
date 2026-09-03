.PHONY: proto test run

proto:
	protoc --go_out=. --go_opt=module=github.com/Qesy/QesyWeb proto/game/v1/game.proto

test:
	go test ./...

run:
	go run ./cmd/web
