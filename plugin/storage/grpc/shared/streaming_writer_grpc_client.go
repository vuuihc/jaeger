package shared

import (
	"context"
	"fmt"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"google.golang.org/grpc"
)

type streamingWriterGRPCClient struct {
	*grpcClient
	streamingWriterClient storage_v1.StreamingSpanWriterPluginClient
	stream                storage_v1.StreamingSpanWriterPlugin_WriteSpanStreamClient
}

func NewStreamingWriterGPRCClient(c *grpc.ClientConn) *streamingWriterGRPCClient {
	return &streamingWriterGRPCClient{
		grpcClient:            NewGRPCClient(c),
		streamingWriterClient: storage_v1.NewStreamingSpanWriterPluginClient(c),
	}
}

func (s *streamingWriterGRPCClient) SpanWriter() spanstore.Writer {
	return s
}

func (s *streamingWriterGRPCClient) WriteSpan(ctx context.Context, span *model.Span) error {
	if s.stream == nil {
		var err error
		s.stream, err = s.streamingWriterClient.WriteSpanStream(ctx)
		if err != nil {
			return fmt.Errorf("plugin error: %w", err)
		}
	}
	return s.stream.Send(&storage_v1.WriteSpanRequest{Span: span})
}
