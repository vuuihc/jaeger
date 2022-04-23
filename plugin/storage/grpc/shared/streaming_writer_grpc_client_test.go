// Copyright (c) 2022 The Jaeger Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shared

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	grpcMocks "github.com/jaegertracing/jaeger/proto-gen/storage_v1/mocks"
)

type grpcStreamWriterClientTest struct {
	client              *streamingWriterGRPCClient
	spanWriter          *grpcMocks.SpanWriterPluginClient
	streamingSpanWriter *grpcMocks.StreamingSpanWriterPluginClient
}

func withStreamingWriterGRPCClient(fn func(r *grpcStreamWriterClientTest)) {
	streamingWriterClient := new(grpcMocks.StreamingSpanWriterPluginClient)
	writerClient := new(grpcMocks.SpanWriterPluginClient)
	r := &grpcStreamWriterClientTest{
		client: &streamingWriterGRPCClient{
			grpcClient: &grpcClient{
				writerClient: writerClient,
			},
			streamingWriterClient: streamingWriterClient,
		},
		spanWriter:          writerClient,
		streamingSpanWriter: streamingWriterClient,
	}
	fn(r)
}

func TestNewStreamingWriterGPRCClient(t *testing.T) {
	sc := NewStreamingWriterGPRCClient(&grpc.ClientConn{})
	assert.NotNil(t, sc.grpcClient)
	assert.NotNil(t, sc.streamingWriterClient)
}

func TestStreamClientWriteSpanStream(t *testing.T) {
	withStreamingWriterGRPCClient(func(r *grpcStreamWriterClientTest) {
		stream := new(grpcMocks.SpanWriterPlugin_WriteSpanStreaminglyClient)
		stream.On("Send", &storage_v1.WriteSpanRequest{
			Span: &mockTraceSpans[0],
		}).Return(nil)
		r.streamingSpanWriter.On("WriteSpanStream", mock.Anything).Return(nil, status.Error(codes.DeadlineExceeded, "")).Once().
			On("WriteSpanStream", mock.Anything).Return(stream, nil).Once()

		err := r.client.SpanWriter().WriteSpan(context.Background(), &mockTraceSpans[0])
		assert.Error(t, err)
		err = r.client.SpanWriter().WriteSpan(context.Background(), &mockTraceSpans[0])
		assert.NoError(t, err)

		stream.On("CloseAndRecv").Return(&storage_v1.WriteSpanResponse{}, nil).Once().
			On("CloseAndRecv").Return(nil, status.Error(codes.DeadlineExceeded, ""))
		r.spanWriter.On("Close", context.Background(), &storage_v1.CloseWriterRequest{}).Return(&storage_v1.CloseWriterResponse{}, nil)

		err = r.client.Close()
		assert.NoError(t, err)
		err = r.client.Close()
		assert.Error(t, err)
	})
}

func TestStreamClientClose(t *testing.T) {
	withStreamingWriterGRPCClient(func(r *grpcStreamWriterClientTest) {
		stream := new(grpcMocks.SpanWriterPlugin_WriteSpanStreaminglyClient)
		stream.On("CloseAndRecv").Return(&storage_v1.WriteSpanResponse{}, nil)
		r.spanWriter.On("Close", context.Background(), &storage_v1.CloseWriterRequest{}).Return(&storage_v1.CloseWriterResponse{}, nil).Once()

		err := r.client.Close()
		assert.NoError(t, err)

		r.spanWriter.On("Close", context.Background(), &storage_v1.CloseWriterRequest{}).Return(nil, status.Error(codes.DeadlineExceeded, ""))
		err = r.client.Close()
		assert.Error(t, err)
	})
}
