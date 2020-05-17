package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/covista/commons/internal/config"
	"github.com/covista/commons/internal/database"
	"github.com/covista/commons/internal/logging"
	"github.com/covista/commons/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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
	ctx         context.Context
	db          *database.Database
	grpcAddress string
	httpAddress string
	grpcServer  *grpc.Server
}

func NewWithInsecureDefaults(ctx context.Context) (*Server, error) {

	cfg := &config.Config{
		GRPC: config.GRPC{
			ListenAddress: "localhost",
			Port:          "5000",
		},
		HTTP: config.HTTP{
			ListenAddress: "localhost",
			Port:          "5001",
		},
		Database: config.Database{
			Host:     "localhost",
			Database: "covid19",
			User:     "covid19",
			Password: "covid19databasepassword",
			Port:     "5434",
		},
	}

	return NewFromConfig(ctx, cfg)
}

func NewFromConfig(ctx context.Context, cfg *config.Config) (*Server, error) {
	grpcAddress := fmt.Sprintf("%s:%s", cfg.GRPC.ListenAddress, cfg.GRPC.Port)
	httpAddress := fmt.Sprintf("%s:%s", cfg.HTTP.ListenAddress, cfg.HTTP.Port)

	db, err := database.NewFromConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", err)
	}

	srv := &Server{
		ctx:         ctx,
		grpcAddress: grpcAddress,
		httpAddress: httpAddress,
		db:          db,
		grpcServer:  grpc.NewServer(),
	}
	proto.RegisterDiagnosisDBServer(srv.grpcServer, srv)

	return srv, nil
}

func (srv *Server) Shutdown() error {
	log := logging.FromContext(srv.ctx)
	log.Info("Shutting down server")

	srv.db.Close()
	return nil
}

func (srv *Server) Done() <-chan struct{} {
	return srv.ctx.Done()
}

func (srv *Server) ServeGRPC() error {
	log := logging.FromContext(srv.ctx)
	lis, err := net.Listen("tcp", srv.grpcAddress)
	if err != nil {
		return fmt.Errorf("Could not listen on %s: %w", srv.grpcAddress, err)
	}
	log.Infof("Serving GRPC on %s", srv.grpcAddress)
	return srv.grpcServer.Serve(lis)
}

func (srv *Server) ServeHTTP() error {
	log := logging.FromContext(srv.ctx)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := proto.RegisterDiagnosisDBHandlerFromEndpoint(srv.ctx, mux, srv.grpcAddress, opts)
	if err != nil {
		return err
	}

	log.Infof("Serving HTTP on %s", srv.httpAddress)
	return http.ListenAndServe(srv.httpAddress, mux)
}

func (srv *Server) AddReport(ctx context.Context, report *proto.Report) (*proto.AddReportResponse, error) {
	ctx = logging.WithLogger(ctx)
	err := srv.db.AddReport(ctx, report)
	if err != nil {
		return &proto.AddReportResponse{
			Error: err.Error(),
		}, nil
	}
	return &proto.AddReportResponse{}, nil
}

func (srv *Server) GetDiagnosisKeys(req *proto.GetKeyRequest, client proto.DiagnosisDB_GetDiagnosisKeysServer) error {
	ctx := logging.WithLogger(srv.ctx)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	results, errchan := srv.db.GetDiagnosisKeys(ctx, req)
	for {
		select {
		case err := <-errchan:
			if err == nil {
				return nil
			}
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
	ctx = logging.WithLogger(ctx)
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
