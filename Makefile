streamer:
	go run cmd/streamer/*.go

client:
	go run cmd/client/*.go

kafkazk:
	docker-compose up
