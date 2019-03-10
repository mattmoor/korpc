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
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"io"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "{{.ProtoImportPath}}"
	impl "{{.ImplImportPath}}"
)

type server struct {}

{{.Implementation}}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// The grpcServer is currently configured to serve h2c traffic by default.
	// To configure credentials or encryption, see: https://grpc.io/docs/guides/auth.html#go
	grpcServer := grpc.NewServer()

	pb.Register{{.Service}}Server(grpcServer, &server{})

	grpcServer.Serve(lis)
}

var _ = errors.New("don't complain about the import")

{{range $val := .UnimplementedMethods}}
{{$val}}
{{end}}
`

	streamSkeleton = `
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
                {{.Body}}
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
	streamErrorBody = "errCh <- errors.New(`{{.}}`)"
)

var (
	tmpl         = template.Must(template.New("entrypoint").Parse(entrypointTemplate))
	streamMethod = template.Must(template.New("stream").Parse(streamSkeleton))
	streamError  = template.Must(template.New("streamError").Parse(streamErrorBody))
)
