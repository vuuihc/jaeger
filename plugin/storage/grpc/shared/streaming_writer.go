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
	"sync"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/pkg/multierror"
	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

var (
	_ spanstore.Writer = (*streamingSpanWriter)(nil)
)

// streamingSpanWriter wraps storage_v1.StreamingSpanWriterPluginClient into spanstore.Writer
type streamingSpanWriter struct {
	client     storage_v1.StreamingSpanWriterPluginClient
	streamPool []storage_v1.StreamingSpanWriterPlugin_WriteSpanStreamClient
	closed     bool
	mu         sync.Mutex
}

func newStreamingSpanWriter(client storage_v1.StreamingSpanWriterPluginClient) *streamingSpanWriter {
	s := &streamingSpanWriter{client: client, mu: sync.Mutex{}, streamPool: make([]storage_v1.StreamingSpanWriterPlugin_WriteSpanStreamClient, 0, 1000)}
	return s
}

// WriteSpan write span into stream
func (s *streamingSpanWriter) WriteSpan(ctx context.Context, span *model.Span) error {
	stream, err := s.getStream(ctx)
	if err != nil {
		return fmt.Errorf("plugin getStream error: %w", err)
	}
	if err := stream.Send(&storage_v1.WriteSpanRequest{Span: span}); err != nil {
		return fmt.Errorf("plugin Send error: %w", err)
	}
	s.putStream(stream)
	return nil
}

func (s *streamingSpanWriter) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	errMu := sync.Mutex{}
	for i := range s.streamPool {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := s.streamPool[i].CloseAndRecv(); err != nil {
				errMu.Lock()
				errs = append(errs, err)
			}
		}(i)
	}
	wg.Wait()
	s.closed = true
	return multierror.Wrap(errs)
}

func (s *streamingSpanWriter) getStream(ctx context.Context) (storage_v1.StreamingSpanWriterPlugin_WriteSpanStreamClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("plugin has closed")
	}
	poolSize := len(s.streamPool)
	if poolSize > 0 {
		stream := s.streamPool[poolSize-1]
		s.streamPool = s.streamPool[:poolSize-1]
		return stream, nil
	}
	return s.client.WriteSpanStream(ctx)
}

func (s *streamingSpanWriter) putStream(stream storage_v1.StreamingSpanWriterPlugin_WriteSpanStreamClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streamPool = append(s.streamPool, stream)
}
