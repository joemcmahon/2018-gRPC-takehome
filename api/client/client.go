package Client

import (
	"context"
	"log"

	pb "github.com/joemcmahon/joe_macmahon_technical_test/api/crawl"
	"google.golang.org/grpc"
)

// CrawlClient encapsulates a connection to the Crawler;
// the RPC methods may be called on it.
type CrawlClient struct {
	conn   *grpc.ClientConn
	client pb.CrawlClient
}

// CrawlSite allows us to control the crawl on a specific site.
func (c *CrawlClient) CrawlSite(ctx context.Context, in *pb.URLRequest, opts ...grpc.CallOption) (*pb.URLState, error) {
	return c.client.CrawlSite(ctx, in, opts...)
}

// CrawlResult allows us to check on the results for a crawl.
func (c *CrawlClient) CrawlResult(ctx context.Context, in *pb.URLRequest, opts ...grpc.CallOption) (*pb.SiteNode, error) {
	return c.client.CrawlResult(ctx, in, opts...)
}

// New takes the gRPC connection data, connects to the server,
// and returns a struct that the client methods can be called on.
func New(serverAddr string, opts ...grpc.DialOption) *CrawlClient {
	conn, client := setup(serverAddr, opts)
	z := CrawlClient{conn: conn, client: client}
	return &z
}

// setup wraps up all the mechanics to create the connection and client so they
// can be saved in the CrawClient.
func setup(serverAddr string, opts []grpc.DialOption) (*grpc.ClientConn, pb.CrawlClient) {
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	client := pb.NewCrawlClient(conn)
	return conn, client
}

// Close exists to allow us to defer the close of the connection.
func (c *CrawlClient) Close() {
	c.conn.Close()
}
