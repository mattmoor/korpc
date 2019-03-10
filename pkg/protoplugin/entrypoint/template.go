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
)
