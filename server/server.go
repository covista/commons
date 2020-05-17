package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/covista/commons/config"
	"github.com/covista/commons/database"
	"github.com/covista/commons/proto"
	"google.golang.org/grpc"
)

func checkConfig(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("Configuration is nil")
	} else if len(cfg.GRPC.Port) == 0 {
		return errors.New("GRPC.Port is empty")
	} else if len(cfg.GRPC.ListenAddress) == 0 {
		return errors.New("GRPC.ListenAddress is empty")
	} else {
		return nil
	}
}

type Server struct {
	db      *database.Database
	address string
}

func NewWithInsecureDefaults(ctx context.Context) (*Server, error) {
	address := "localhost:5000"

	db, err := database.NewWithInsecureDefaults(ctx)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", err)
	}

	srv := &Server{
		address: address,
		db:      db,
	}

	return srv, nil
}

func NewFromConfig(ctx context.Context, cfg *config.Config) (*Server, error) {
	address := fmt.Sprintf("%s:%s", cfg.GRPC.ListenAddress, cfg.GRPC.Port)

	db, err := database.NewFromConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", err)
	}

	srv := &Server{
		address: address,
		db:      db,
	}

	return srv, nil
}

func (srv *Server) Shutdown() error {
	srv.db.Close()
	return nil
}

func (srv *Server) ServeGRPC() error {
	lis, err := net.Listen("tcp", srv.address)
	if err != nil {
		return fmt.Errorf("Could not listen on %s: %w", srv.address, err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterDiagnosisDBServer(grpcServer, srv)
	return grpcServer.Serve(lis)
}

func (srv *Server) AddReport(ctx context.Context, report *proto.Report) (*proto.AddReportResponse, error) {
	err := srv.db.AddReport(ctx, report)
	if err != nil {
		return &proto.AddReportResponse{
			Error: err.Error(),
		}, nil
	}
	return &proto.AddReportResponse{}, nil
}

func (srv *Server) GetDiagnosisKeys(req *proto.GetKeyRequest, client proto.DiagnosisDB_GetDiagnosisKeysServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	results, errchan := srv.db.GetDiagnosisKeys(ctx, req)
	for {
		select {
		case err := <-errchan:
			serr := client.Send(&proto.GetDiagnosisKeyResponse{
				Error: err.Error(),
			})
			if serr != nil {
				return fmt.Errorf("Could not send error (%w): %w", err, serr)
			}
		case tstek := <-results:
			if tstek == nil {
				return nil
			}
			serr := client.Send(&proto.GetDiagnosisKeyResponse{
				Record: tstek,
			})
			if serr != nil {
				return fmt.Errorf("Could not send: %w", serr)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (srv *Server) GetAuthorizationToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error) {
	one_time_auth_key, err := srv.db.CreateAuthorizationKey(ctx, req)
	if err != nil {
		return &proto.TokenResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.TokenResponse{
		AuthorizationKey: one_time_auth_key,
	}, nil
}
