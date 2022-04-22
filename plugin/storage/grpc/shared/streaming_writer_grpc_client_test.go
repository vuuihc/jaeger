package shared

import (
	"context"
	"testing"

	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	grpcMocks "github.com/jaegertracing/jaeger/proto-gen/storage_v1/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type grpcStreamWriterClientTest struct {
	client              *streamingWriterGRPCClient
	streamingSpanWriter *grpcMocks.StreamingSpanWriterPluginClient
	stream              *grpcMocks.StreamingSpanWriterPlugin_WriteSpanStreamClient
}

func withStreamingWriterGRPCClient(fn func(r *grpcStreamWriterClientTest)) {
	streamingSpanWriter := new(grpcMocks.StreamingSpanWriterPluginClient)
	stream := new(grpcMocks.StreamingSpanWriterPlugin_WriteSpanStreamClient)
	r := &grpcStreamWriterClientTest{
		client: &streamingWriterGRPCClient{
			streamingWriterClient: streamingSpanWriter,
			stream:                stream,
		},
		streamingSpanWriter: streamingSpanWriter,
		stream:              stream,
	}
	fn(r)
}

func TestStreamClientWriteSpan(t *testing.T) {
	withStreamingWriterGRPCClient(func(r *grpcStreamWriterClientTest) {
		r.streamingSpanWriter.On("WriteSpanStream", mock.Anything).Return(r.stream, nil)
		r.stream.On("Send", &storage_v1.WriteSpanRequest{
			Span: &mockTraceSpans[0],
		}).Return(nil)
		err := r.client.WriteSpan(context.Background(), &mockTraceSpans[0])
		assert.NoError(t, err)
	})
}
