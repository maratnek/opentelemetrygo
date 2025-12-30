package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"opentelemetry/proto"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type transServer struct {
	proto.UnimplementedTransServiceServer
}

func (s *transServer) Create(ctx context.Context, req *proto.CreateRequest) (*proto.CreateResponse, error) {
	fmt.Println("Create request")
	tracer := otel.Tracer("trans-server")
	_, span := tracer.Start(ctx, "TransService.Create")
	defer span.End()

	span.SetAttributes(attribute.String("request.message", req.Message))

	if len(req.Message) == 0 {
		span.RecordError(status.Error(codes.InvalidArgument, "message is empty"))
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	resp := &proto.CreateResponse{
		Message: "Created: " + req.Message,
	}

	log.Printf("Received message: %s", req.Message)
	return resp, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	shutdown := initOTel(ctx)
	defer shutdown(context.Background())

	url := ":50051"
	lis, err := net.Listen("tcp", url)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()), // <-- новый способ (tracing + metrics)
	)

	proto.RegisterTransServiceServer(grpcServer, &transServer{})

	go func() {
		log.Println("gRPC server starting on ", url)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("server error: %v", err)
		}
		log.Println("Server started!")
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
	log.Println("Server stopped gracefully")
}
