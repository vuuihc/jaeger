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

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

var (
	_ spanstore.Writer = (*streamingSpanWriter)(nil)
)

// streamingSpanWriter wraps storage_v1.StreamingSpanWriterPluginClient into spanstore.Writer
type streamingSpanWriter struct {
	client storage_v1.StreamingSpanWriterPluginClient
	stream storage_v1.StreamingSpanWriterPlugin_WriteSpanStreamClient
}

// WriteSpan write span into stream
func (s *streamingSpanWriter) WriteSpan(ctx context.Context, span *model.Span) error {
	if s.stream == nil {
		var err error
		s.stream, err = s.client.WriteSpanStream(ctx)
		if err != nil {
			return fmt.Errorf("plugin WriteSpanStream error: %w", err)
		}
	}
	if err := s.stream.Send(&storage_v1.WriteSpanRequest{Span: span}); err != nil {
		s.stream = nil
		return fmt.Errorf("plugin Send error: %w", err)
	}
	return nil
}

func (s *streamingSpanWriter) Close() error {
	if s.stream != nil {
		if _, err := s.stream.CloseAndRecv(); err != nil {
			return fmt.Errorf("plugin error: %w", err)
		}
	}
	return nil
}
