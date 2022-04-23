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

	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	grpcMocks "github.com/jaegertracing/jaeger/proto-gen/storage_v1/mocks"
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
