package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/jaegertracing/jaeger/proto-gen/storage_v1"
	"google.golang.org/grpc"
)

// Ensure plugin.GRPCPlugin API match.
var _ plugin.GRPCPlugin = (*StorageStreamingWriterGRPCPlugin)(nil)

// StorageStreamingWriterGRPCPlugin is the implementation of plugin.GRPCPlugin.
type StorageStreamingWriterGRPCPlugin struct {
	plugin.Plugin
	// Concrete implementation, This is only used for plugins that are written in Go.
	Impl        StoragePlugin
	ArchiveImpl ArchiveStoragePlugin
}

// RegisterHandlers registers the plugin with the server
func (p *StorageStreamingWriterGRPCPlugin) RegisterHandlers(s *grpc.Server) error {
	server := &grpcServer{
		Impl:        p.Impl,
		ArchiveImpl: p.ArchiveImpl,
	}
	storage_v1.RegisterSpanReaderPluginServer(s, server)
	storage_v1.RegisterSpanWriterPluginServer(s, server)
	storage_v1.RegisterArchiveSpanReaderPluginServer(s, server)
	storage_v1.RegisterArchiveSpanWriterPluginServer(s, server)
	storage_v1.RegisterPluginCapabilitiesServer(s, server)
	storage_v1.RegisterDependenciesReaderPluginServer(s, server)
	storage_v1.RegisterStreamingSpanWriterPluginServer(s, server)
	return nil
}

// GRPCServer implements plugin.GRPCPlugin. It is used by go-plugin to create a grpc plugin server.
func (p *StorageStreamingWriterGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	return p.RegisterHandlers(s)
}

// GRPCClient implements plugin.GRPCPlugin. It is used by go-plugin to create a grpc plugin client.
func (*StorageStreamingWriterGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return NewStreamingWriterGPRCClient(c), nil
}
