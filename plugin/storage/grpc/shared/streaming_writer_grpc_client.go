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
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	"github.com/jaegertracing/jaeger/storage/spanstore"
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

func (s *streamingWriterGRPCClient) Close() error {
	if s.stream != nil {
		if _, err := s.stream.CloseAndRecv(); err != nil {
			return fmt.Errorf("plugin error: %w", err)
		}
	}
	if err := s.grpcClient.Close(); err != nil && status.Code(err) != codes.Unimplemented {
		return fmt.Errorf("plugin error: %w", err)
	}
	return nil
}
