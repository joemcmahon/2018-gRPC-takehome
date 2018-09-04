package common

import (
	"log"

	pb "github.com/joemcmahon/joe_macmahon_technical_test/api/crawl"
	"github.com/joemcmahon/joe_macmahon_technical_test/testdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Connect(tls bool, caFile string, hostOverride string, serverAddr string) pb.CrawlClient {
	var opts []grpc.DialOption
	if tls {
		if caFile == "" {
			caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, hostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	return pb.NewCrawlClient(conn)
}
