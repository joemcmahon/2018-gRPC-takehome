all:
	protoc -I api/ api/crawl.proto --go_out=plugins=grpc:api/crawl
	go build server.go
	go build crawl.go

run:    all
	killall server || echo "No server to kill"
	TESTING=$(TESTING) ./server &

mock:   all
	killall server || echo "No server to kill"
	TESTING=$(TESTING) ./server -mock &

test:	all
	$(GOPATH)/bin/ginkgo -r -v
