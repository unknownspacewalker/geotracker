package historyclient

import (
	"context"
	"fmt"
	log2 "log"

	"gitlab.com/spacewalker/geotracker/internal/app/location/core/port"
	"gitlab.com/spacewalker/geotracker/internal/pkg/errpack"
	"gitlab.com/spacewalker/geotracker/internal/pkg/geo"
	"gitlab.com/spacewalker/geotracker/internal/pkg/log"
	"gitlab.com/spacewalker/geotracker/internal/pkg/middleware"
	pb "gitlab.com/spacewalker/geotracker/pkg/api/proto/v1/history"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCClient struct {
	addr   string
	logger log.Logger
}

// NewGRPCClient TODO: add description
func NewGRPCClient(addr string, logger log.Logger) port.HistoryClient {
	if logger == nil {
		log2.Panic("logger must not be nil")
	}

	return &GRPCClient{
		addr:   addr,
		logger: logger,
	}
}

// AddRecord TODO: add description
func (c GRPCClient) AddRecord(ctx context.Context, req port.HistoryClientAddRecordRequest) (port.HistoryClientAddRecordResponse, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(
			middleware.TracingUnaryClientInterceptor(c.logger),
			middleware.LoggerUnaryClientInterceptor(c.logger),
		),
	}

	conn, err := grpc.Dial(c.addr, opts...)
	if err != nil {
		return port.HistoryClientAddRecordResponse{}, fmt.Errorf("%w: %v", errpack.ErrInternalError, err)
	}
	defer conn.Close()

	client := pb.NewHistoryClient(conn)

	res, err := client.AddRecord(ctx, &pb.AddRecordRequest{
		UserId: int32(req.UserID),
		A: &pb.Point{
			Longitude: req.A.Longitude(),
			Latitude:  req.A.Latitude(),
		},
		B: &pb.Point{
			Longitude: req.B.Longitude(),
			Latitude:  req.B.Latitude(),
		},
		Timestamp: timestamppb.New(req.Timestamp),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return port.HistoryClientAddRecordResponse{}, fmt.Errorf("%w: %v", errpack.ErrInternalError, err)
		}
		switch st.Code() {
		case codes.InvalidArgument:
			return port.HistoryClientAddRecordResponse{}, fmt.Errorf("%w", errpack.ErrInvalidArgument)
		default:
			return port.HistoryClientAddRecordResponse{}, fmt.Errorf("%w: %v", errpack.ErrInternalError, err)
		}
	}

	return port.HistoryClientAddRecordResponse{
		UserID:    int(res.UserId),
		A:         geo.Point{res.A.Longitude, res.A.Latitude},
		B:         geo.Point{res.B.Longitude, res.B.Latitude},
		Timestamp: res.Timestamp.AsTime(),
	}, nil
}
