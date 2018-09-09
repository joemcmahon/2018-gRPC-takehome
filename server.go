package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	pb "github.com/joemcmahon/joe_macmahon_technical_test/api/crawl"
	"github.com/joemcmahon/joe_macmahon_technical_test/api/server"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/fetcher"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/test/mock_fetcher"
	"github.com/joemcmahon/joe_macmahon_technical_test/testdata"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("tls_cert_file", "", "The TLS cert file")
	keyFile  = flag.String("tls_key_file", "", "The TLS key file")
	port     = flag.Int("port", 10000, "The server port")
	debug    = flag.Bool("debug", false, "Turn on server debug")
	mock     = flag.Bool("mock", false, "Use the mock fetcher for testing")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = testdata.Path("localhost.cert")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("localhost.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	if debugging() {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("starting server")
	grpcServer := grpc.NewServer(opts...)
	log.Debug("registering crawler")
	if *mock {
		f := MockFetcher.New()
		pb.RegisterCrawlServer(grpcServer, Server.New(f))
	} else {
		f := Fetcher.New()
		pb.RegisterCrawlServer(grpcServer, Server.New(f))
	}
	log.Debug("ready")
	grpcServer.Serve(lis)
	log.Debug("server terminated")
}

func debugging() bool {
	return *debug || os.Getenv("TESTING") != ""
}
