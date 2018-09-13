all:
	go build server.go
	go build crawl.go

run:	all
	killall server || echo "No server to kill"
	./server &

mock:	all
	killall server || echo "No server to kill"
	./server -mock &
