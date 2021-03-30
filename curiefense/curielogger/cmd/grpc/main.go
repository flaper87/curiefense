package main

import (
	"context"
	"fmt"
	"github.com/curiefense/curiefense/curielogger/pkg"
	"github.com/curiefense/curiefense/curielogger/pkg/drivers"
	als "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	GRPC_LISTENER = `CURIELOGGER_GRPC_LISTEN`
)

func main() {
	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			pkg.NewConfig,
			drivers.InitDrivers,
			pkg.NewMetrics,
			pkg.NewLogSender,
			newGrpcSrv,
		),
		fx.Invoke(grpcInit),
	)
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}

func grpcInit(srv *grpcServer, v *viper.Viper) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%s", v.GetString(GRPC_LISTENER)))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	log.Printf("GRPC server listening on %v", v.GetString(GRPC_LISTENER))
	als.RegisterAccessLogServiceServer(s, srv)
	if err := s.Serve(sock); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
