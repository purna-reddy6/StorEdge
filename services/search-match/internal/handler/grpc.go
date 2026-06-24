package handler

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/storedge/storedge/services/search-match/internal/service"
)

// NewGRPCServer creates a gRPC server wiring the matching service.
// Full protobuf codegen would replace this stub in production.
// For MVP, the HTTP REST API is the primary interface.
func NewGRPCServer(matchingSvc *service.MatchingService, bookingSvc *service.BookingService, logger *zap.Logger) *grpc.Server {
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(16 * 1024 * 1024), // 16MB
		grpc.MaxSendMsgSize(16 * 1024 * 1024),
	)
	// gRPC service registration would go here after proto codegen:
	// pb.RegisterWarehouseMatchServiceServer(server, &warehouseGRPCHandler{matchingSvc, bookingSvc, logger})
	logger.Info("gRPC server configured (proto codegen pending)")
	return server
}
