// Copyright 2019 Matt Moore
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entrypoint

import (
	"text/template"
)

type options struct {
	ProtoImportPath      string
	ImplImportPath       string
	Service              string
	Implementation       string
	UnimplementedMethods []string
}

const (
	entrypointTemplate = `package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"

	pb "{{.ProtoImportPath}}"
	impl "{{.ImplImportPath}}"
)

type server struct {
  pb.Unimplemented{{.Service}}Server
}

{{.Implementation}}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "probe" {
		probe()
		return
	}

	// Prior to this, exporters should be registered via:
	// func init() {
	//    view.RegisterExporter(...)
	// }

	// Register the views to collect server request count.
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// The grpcServer is currently configured to serve h2c traffic by default.
	// To configure credentials or encryption, see: https://grpc.io/docs/guides/auth.html#go
	grpcServer := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))

	pb.Register{{.Service}}Server(grpcServer, &server{})
	healthpb.RegisterHealthServer(grpcServer, &health{})

	grpcServer.Serve(lis)
}

// Based on github.com/grpc-ecosystem/grpc-health-probe
func probe() {
	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%s", os.Getenv("PORT")), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: "{{.Service}}"})
	if err != nil {
		log.Fatalf("Error health checking: %v", err)
	}
	log.Printf("Health check: %#v", resp)
}

type health struct {}

func (h *health) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (h *health) Watch(*healthpb.HealthCheckRequest, healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "korpc does not implement Watch")
}

// Don't complain about the import
var _ = io.EOF
`

	streamInSkeleton = `
func {{.Receiver}}{{.Name}}(stream pb.{{.Service}}_{{.Method}}Server) error {
	input := make(chan *pb.{{.RequestType}})

	errCh := make(chan error)

	go func() {
		defer close(input)
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errCh <- err
				return
			}
			input <- req
		}
	}()

	go func() {
		defer close(errCh)
		resp, err := impl.Impl(stream.Context(), input)
		if err != nil {
			errCh <- err
		} else if err := stream.SendAndClose(resp); err != nil {
			errCh <- err
		}
	}()

	return <-errCh
}
`

	streamOutSkeleton = `
func {{.Receiver}}{{.Name}}(input *pb.{{.RequestType}}, stream pb.{{.Service}}_{{.Method}}Server) error {
	output := make(chan *pb.{{.ResponseType}})
	errCh := make(chan error)

	go func() {
		defer close(output)
		errCh <- impl.Impl(stream.Context(), input, output)
	}()

	go func() {
		defer close(errCh)
		for {
			select {
			case resp, ok := <-output:
				if !ok {
					// Quit when the output channel closes.
					return
				}
				if err := stream.Send(resp); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return <-errCh
}
`

	streamInOutSkeleton = `
func {{.Receiver}}{{.Name}}(stream pb.{{.Service}}_{{.Method}}Server) error {
	input := make(chan *pb.{{.RequestType}})
	output := make(chan *pb.{{.ResponseType}})

	errCh := make(chan error)

	go func() {
		defer close(input)
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errCh <- err
				return
			}
			input <- req
		}
	}()

	go func() {
		defer close(output)
		errCh <- impl.Impl(stream.Context(), input, output)
	}()

	go func() {
		defer close(errCh)
		for {
			select {
			case resp, ok := <-output:
				if !ok {
					// Quit when the output channel closes.
					return
				}
				if err := stream.Send(resp); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return <-errCh
}
`
)

var (
	tmpl              = template.Must(template.New("entrypoint").Parse(entrypointTemplate))
	streamInOutMethod = template.Must(template.New("stream-in-out").Parse(streamInOutSkeleton))
	streamInMethod    = template.Must(template.New("stream-in").Parse(streamInSkeleton))
	streamOutMethod   = template.Must(template.New("stream-out").Parse(streamOutSkeleton))
)
