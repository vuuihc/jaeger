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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	grpcMocks "github.com/jaegertracing/jaeger/proto-gen/storage_v1/mocks"
)

type streamingSpanWriterTest struct {
	client              *streamingSpanWriter
	streamingSpanWriter *grpcMocks.StreamingSpanWriterPluginClient
}

func withStreamingWriterGRPCClient(fn func(r *streamingSpanWriterTest)) {
	streamingWriterClient := new(grpcMocks.StreamingSpanWriterPluginClient)
	r := &streamingSpanWriterTest{
		client:              newStreamingSpanWriter(streamingWriterClient),
		streamingSpanWriter: streamingWriterClient,
	}
	fn(r)
}

func TestStreamClientWriteSpanStream(t *testing.T) {
	withStreamingWriterGRPCClient(func(r *streamingSpanWriterTest) {
		stream := new(grpcMocks.StreamingSpanWriterPlugin_WriteSpanStreamClient)
		stream.On("Send", &storage_v1.WriteSpanRequest{Span: &mockTraceSpans[0]}).Return(io.EOF).Once().
			On("Send", &storage_v1.WriteSpanRequest{Span: &mockTraceSpans[0]}).Return(nil).Once()
		r.streamingSpanWriter.On("WriteSpanStream", mock.Anything).Return(nil, status.Error(codes.DeadlineExceeded, "timeout")).Once().
			On("WriteSpanStream", mock.Anything).Return(stream, nil)

		err := r.client.WriteSpan(context.Background(), &mockTraceSpans[0])
		assert.ErrorContains(t, err, "timeout")
		err = r.client.WriteSpan(context.Background(), &mockTraceSpans[0])
		assert.ErrorContains(t, err, "EOF")
		err = r.client.WriteSpan(context.Background(), &mockTraceSpans[0])
		assert.NoError(t, err)
	})
}

func TestStreamClientClose(t *testing.T) {
	withStreamingWriterGRPCClient(func(r *streamingSpanWriterTest) {
		stream := new(grpcMocks.StreamingSpanWriterPlugin_WriteSpanStreamClient)
		stream.On("CloseAndRecv").Return(&storage_v1.WriteSpanResponse{}, nil).Once().
			On("CloseAndRecv").Return(nil, status.Error(codes.DeadlineExceeded, ""))
		// r.client.streamPool = append(r.client.streamPool, stream)
		r.client.streamPool <- stream

		err := r.client.Close()
		assert.NoError(t, err)
		err = r.client.Close()
		assert.Error(t, err)
	})
}
