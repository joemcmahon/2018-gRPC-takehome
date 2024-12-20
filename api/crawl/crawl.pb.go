// Code generated by protoc-gen-go. DO NOT EDIT.
// source: crawl.proto

package crawl

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type URLRequestCommand int32

const (
	// URLs in STOPPED, NONE, or DONE may be started.
	URLRequest_START URLRequestCommand = 0
	// URLs in CRAWLING, STOPPED, or DONE may be stopped.
	URLRequest_STOP URLRequestCommand = 1
	// URLs in any state may be checked.
	URLRequest_CHECK URLRequestCommand = 2
	// We have to stop, show, then start again.
	URLRequest_SHOW URLRequestCommand = 3
)

var URLRequestCommand_name = map[int32]string{
	0: "START",
	1: "STOP",
	2: "CHECK",
	3: "SHOW",
}
var URLRequestCommand_value = map[string]int32{
	"START": 0,
	"STOP":  1,
	"CHECK": 2,
	"SHOW":  3,
}

func (x URLRequestCommand) String() string {
	return proto.EnumName(URLRequestCommand_name, int32(x))
}
func (URLRequestCommand) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_crawl_c886f2d72542114f, []int{0, 0}
}

type URLState_Status int32

const (
	URLState_STOPPED URLState_Status = 0
	// START for a STOPPED URL resumes the crawl.
	// STOP for a STOPPED URL does nothing.
	URLState_RUNNING URLState_Status = 1
	// Once it completes the crawl, it switches
	// the URL's state to DONE. START for a
	// CRAWLING URL is a no-op. STOP for a CRAWLING
	// URL saves the URL's state and sets it to STOPPED.
	URLState_UNKNOWN URLState_Status = 2
)

var URLState_Status_name = map[int32]string{
	0: "STOPPED",
	1: "RUNNING",
	2: "UNKNOWN",
}
var URLState_Status_value = map[string]int32{
	"STOPPED": 0,
	"RUNNING": 1,
	"UNKNOWN": 2,
}

func (x URLState_Status) String() string {
	return proto.EnumName(URLState_Status_name, int32(x))
}
func (URLState_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_crawl_c886f2d72542114f, []int{1, 0}
}

// URLRequest defines the outgoing request.
// We can provide a URL and the state we want the client
// to put it in.
type URLRequest struct {
	URL                  string            `protobuf:"bytes,1,opt,name=URL,proto3" json:"URL,omitempty"`
	State                URLRequestCommand `protobuf:"varint,2,opt,name=state,proto3,enum=crawl.URLRequestCommand" json:"state,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *URLRequest) Reset()         { *m = URLRequest{} }
func (m *URLRequest) String() string { return proto.CompactTextString(m) }
func (*URLRequest) ProtoMessage()    {}
func (*URLRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_crawl_c886f2d72542114f, []int{0}
}
func (m *URLRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_URLRequest.Unmarshal(m, b)
}
func (m *URLRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_URLRequest.Marshal(b, m, deterministic)
}
func (dst *URLRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_URLRequest.Merge(dst, src)
}
func (m *URLRequest) XXX_Size() int {
	return xxx_messageInfo_URLRequest.Size(m)
}
func (m *URLRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_URLRequest.DiscardUnknown(m)
}

var xxx_messageInfo_URLRequest proto.InternalMessageInfo

func (m *URLRequest) GetURL() string {
	if m != nil {
		return m.URL
	}
	return ""
}

func (m *URLRequest) GetState() URLRequestCommand {
	if m != nil {
		return m.State
	}
	return URLRequest_START
}

// URLState reports the crawl status ONLY of a URL.
type URLState struct {
	Status               URLState_Status `protobuf:"varint,1,opt,name=status,proto3,enum=crawl.URLState_Status" json:"status,omitempty"`
	Message              string          `protobuf:"bytes,2,opt,name=Message,proto3" json:"Message,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *URLState) Reset()         { *m = URLState{} }
func (m *URLState) String() string { return proto.CompactTextString(m) }
func (*URLState) ProtoMessage()    {}
func (*URLState) Descriptor() ([]byte, []int) {
	return fileDescriptor_crawl_c886f2d72542114f, []int{1}
}
func (m *URLState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_URLState.Unmarshal(m, b)
}
func (m *URLState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_URLState.Marshal(b, m, deterministic)
}
func (dst *URLState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_URLState.Merge(dst, src)
}
func (m *URLState) XXX_Size() int {
	return xxx_messageInfo_URLState.Size(m)
}
func (m *URLState) XXX_DiscardUnknown() {
	xxx_messageInfo_URLState.DiscardUnknown(m)
}

var xxx_messageInfo_URLState proto.InternalMessageInfo

func (m *URLState) GetStatus() URLState_Status {
	if m != nil {
		return m.Status
	}
	return URLState_STOPPED
}

func (m *URLState) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

// SiteNode is returned in response to a STATUS request.
// It returns a tree of sitenodes found under the current
// URL (which may recursively contain more SiteNodes).
// If no URL is supplied, all the SiteNodes the crawler
// knows about are returned as the children of a SiteNode
// with the siteURL "all://".
type SiteNode struct {
	SiteURL              string   `protobuf:"bytes,1,opt,name=siteURL,proto3" json:"siteURL,omitempty"`
	TreeString           string   `protobuf:"bytes,2,opt,name=treeString,proto3" json:"treeString,omitempty"`
	Status               string   `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SiteNode) Reset()         { *m = SiteNode{} }
func (m *SiteNode) String() string { return proto.CompactTextString(m) }
func (*SiteNode) ProtoMessage()    {}
func (*SiteNode) Descriptor() ([]byte, []int) {
	return fileDescriptor_crawl_c886f2d72542114f, []int{2}
}
func (m *SiteNode) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SiteNode.Unmarshal(m, b)
}
func (m *SiteNode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SiteNode.Marshal(b, m, deterministic)
}
func (dst *SiteNode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SiteNode.Merge(dst, src)
}
func (m *SiteNode) XXX_Size() int {
	return xxx_messageInfo_SiteNode.Size(m)
}
func (m *SiteNode) XXX_DiscardUnknown() {
	xxx_messageInfo_SiteNode.DiscardUnknown(m)
}

var xxx_messageInfo_SiteNode proto.InternalMessageInfo

func (m *SiteNode) GetSiteURL() string {
	if m != nil {
		return m.SiteURL
	}
	return ""
}

func (m *SiteNode) GetTreeString() string {
	if m != nil {
		return m.TreeString
	}
	return ""
}

func (m *SiteNode) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func init() {
	proto.RegisterType((*URLRequest)(nil), "crawl.URLRequest")
	proto.RegisterType((*URLState)(nil), "crawl.URLState")
	proto.RegisterType((*SiteNode)(nil), "crawl.SiteNode")
	proto.RegisterEnum("crawl.URLRequestCommand", URLRequestCommand_name, URLRequestCommand_value)
	proto.RegisterEnum("crawl.URLState_Status", URLState_Status_name, URLState_Status_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// CrawlClient is the client API for Crawl service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CrawlClient interface {
	// Because we're calling the client from our CLI, we
	// want the CrawlSite API to make a single request
	// and wait for the response. This API lets us start,
	// stop, or check the status of a URL
	CrawlSite(ctx context.Context, in *URLRequest, opts ...grpc.CallOption) (*URLState, error)
	// Checks the current status of a crawl and returns
	// the tree as it stands.
	CrawlResult(ctx context.Context, in *URLRequest, opts ...grpc.CallOption) (Crawl_CrawlResultClient, error)
}

type crawlClient struct {
	cc *grpc.ClientConn
}

func NewCrawlClient(cc *grpc.ClientConn) CrawlClient {
	return &crawlClient{cc}
}

func (c *crawlClient) CrawlSite(ctx context.Context, in *URLRequest, opts ...grpc.CallOption) (*URLState, error) {
	out := new(URLState)
	err := c.cc.Invoke(ctx, "/crawl.Crawl/CrawlSite", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlClient) CrawlResult(ctx context.Context, in *URLRequest, opts ...grpc.CallOption) (Crawl_CrawlResultClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Crawl_serviceDesc.Streams[0], "/crawl.Crawl/CrawlResult", opts...)
	if err != nil {
		return nil, err
	}
	x := &crawlCrawlResultClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Crawl_CrawlResultClient interface {
	Recv() (*SiteNode, error)
	grpc.ClientStream
}

type crawlCrawlResultClient struct {
	grpc.ClientStream
}

func (x *crawlCrawlResultClient) Recv() (*SiteNode, error) {
	m := new(SiteNode)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CrawlServer is the server API for Crawl service.
type CrawlServer interface {
	// Because we're calling the client from our CLI, we
	// want the CrawlSite API to make a single request
	// and wait for the response. This API lets us start,
	// stop, or check the status of a URL
	CrawlSite(context.Context, *URLRequest) (*URLState, error)
	// Checks the current status of a crawl and returns
	// the tree as it stands.
	CrawlResult(*URLRequest, Crawl_CrawlResultServer) error
}

func RegisterCrawlServer(s *grpc.Server, srv CrawlServer) {
	s.RegisterService(&_Crawl_serviceDesc, srv)
}

func _Crawl_CrawlSite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlServer).CrawlSite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crawl.Crawl/CrawlSite",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlServer).CrawlSite(ctx, req.(*URLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crawl_CrawlResult_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(URLRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CrawlServer).CrawlResult(m, &crawlCrawlResultServer{stream})
}

type Crawl_CrawlResultServer interface {
	Send(*SiteNode) error
	grpc.ServerStream
}

type crawlCrawlResultServer struct {
	grpc.ServerStream
}

func (x *crawlCrawlResultServer) Send(m *SiteNode) error {
	return x.ServerStream.SendMsg(m)
}

var _Crawl_serviceDesc = grpc.ServiceDesc{
	ServiceName: "crawl.Crawl",
	HandlerType: (*CrawlServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CrawlSite",
			Handler:    _Crawl_CrawlSite_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CrawlResult",
			Handler:       _Crawl_CrawlResult_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "crawl.proto",
}

func init() { proto.RegisterFile("crawl.proto", fileDescriptor_crawl_c886f2d72542114f) }

var fileDescriptor_crawl_c886f2d72542114f = []byte{
	// 321 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x51, 0xc1, 0x4e, 0x02, 0x31,
	0x14, 0xdc, 0x82, 0x0b, 0xec, 0x23, 0xc1, 0xfa, 0x0e, 0x04, 0x3d, 0x18, 0xd2, 0x13, 0xa7, 0x45,
	0x21, 0x7e, 0x80, 0x41, 0x22, 0x06, 0x2c, 0xa4, 0x65, 0xc3, 0xc5, 0xcb, 0x0a, 0x0d, 0x21, 0x01,
	0x56, 0xb6, 0x25, 0xfe, 0x80, 0xfe, 0xb7, 0x69, 0x77, 0x71, 0x8d, 0x26, 0xde, 0x3a, 0xd3, 0x79,
	0x33, 0x9d, 0x57, 0xa8, 0x2f, 0xd3, 0xf8, 0x7d, 0x1b, 0xbe, 0xa5, 0x89, 0x49, 0xd0, 0x77, 0x80,
	0x7d, 0x10, 0x80, 0x48, 0x4c, 0x84, 0x3a, 0x1c, 0x95, 0x36, 0x48, 0xa1, 0x1c, 0x89, 0x49, 0x8b,
	0xb4, 0x49, 0x27, 0x10, 0xf6, 0x88, 0x5d, 0xf0, 0xb5, 0x89, 0x8d, 0x6a, 0x95, 0xda, 0xa4, 0xd3,
	0xe8, 0x5d, 0x86, 0x99, 0x49, 0x31, 0x13, 0x2e, 0x93, 0xdd, 0x2e, 0xde, 0xaf, 0x44, 0xa6, 0x63,
	0x7d, 0xa8, 0xe6, 0x0c, 0x06, 0xe0, 0xcb, 0xf9, 0xbd, 0x98, 0x53, 0x0f, 0x6b, 0x70, 0x26, 0xe7,
	0xd3, 0x19, 0x25, 0x96, 0x1c, 0x8c, 0x86, 0x83, 0x31, 0x2d, 0x39, 0x72, 0x34, 0x5d, 0xd0, 0x32,
	0xfb, 0x24, 0x50, 0x8b, 0xc4, 0x44, 0x5a, 0x07, 0x0c, 0xa1, 0x62, 0xad, 0x8e, 0xda, 0xbd, 0xa3,
	0xd1, 0x6b, 0x16, 0x99, 0x4e, 0x10, 0x4a, 0x77, 0x2b, 0x72, 0x15, 0xb6, 0xa0, 0xfa, 0xac, 0xb4,
	0x8e, 0xd7, 0xd9, 0x23, 0x03, 0x71, 0x82, 0xac, 0x0b, 0x95, 0x4c, 0x8b, 0x75, 0xa8, 0xda, 0xfc,
	0xd9, 0xf0, 0x81, 0x7a, 0x16, 0x88, 0x88, 0xf3, 0x27, 0xfe, 0x48, 0x89, 0x05, 0x11, 0x1f, 0xf3,
	0xe9, 0x82, 0xd3, 0x12, 0x7b, 0x81, 0x9a, 0xdc, 0x18, 0xc5, 0x93, 0x95, 0xb2, 0xb6, 0x7a, 0x63,
	0x54, 0xb1, 0x8f, 0x13, 0xc4, 0x6b, 0x00, 0x93, 0x2a, 0x25, 0x4d, 0xba, 0xd9, 0xaf, 0xf3, 0xcc,
	0x1f, 0x0c, 0x36, 0xbf, 0x0b, 0x94, 0xdd, 0x5d, 0x8e, 0x7a, 0x07, 0xf0, 0x07, 0xb6, 0x09, 0xde,
	0x42, 0xe0, 0x0e, 0x36, 0x0b, 0x2f, 0xfe, 0xac, 0xf4, 0xea, 0xfc, 0x57, 0x63, 0xe6, 0xe1, 0x1d,
	0xd4, 0xdd, 0x88, 0x50, 0xfa, 0xb8, 0x35, 0xff, 0x0d, 0x9d, 0x0a, 0x30, 0xef, 0x86, 0xbc, 0x56,
	0xdc, 0x6f, 0xf7, 0xbf, 0x02, 0x00, 0x00, 0xff, 0xff, 0x46, 0x41, 0x8d, 0x26, 0xfc, 0x01, 0x00,
	0x00,
}
