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

package scaffold

import (
	"text/template"
)

type options struct {
	Package         string
	ProtoImportPath string
	Body            string
}

const (
	scaffoldTemplate = `package {{.Package}}

import (
	"context"
	"errors"

	pb "{{.ProtoImportPath}}"
)

{{.Body}}
`

	// TODO(mattmoor): Right now these assume that the request/response types
	// come from the same proto where the method is defined.  While this doesn't
	// seem an outrageous assumption for early prototyping, it is something a
	// proper solution would address.
	unaryErrorBody = "return nil, errors.New(`{{.}}`)"
	unarySkeleton  = `
func {{.Receiver}}{{.Name}}(ctx context.Context, req *pb.{{.RequestType}}) (*pb.{{.ResponseType}}, error) {
	{{.Body}}
}
`

	streamInOutErrorBody = `for {
		select {
		case _, ok := <-req:
			if !ok {
				return errors.New("{{.}}")
			}
		}
	}
`

	streamInErrorBody = `for {
		select {
		case _, ok := <-req:
			if !ok {
				return nil, errors.New("{{.}}")
			}
		}
	}
`
	streamOutErrorBody  = "return errors.New(`{{.}}`)"
	streamInOutSkeleton = `
func {{.Receiver}}{{.Name}}(ctx context.Context, req <-chan *pb.{{.RequestType}}, resp chan *pb.{{.ResponseType}}) error {
	{{.Body}}
}
`
	streamInSkeleton = `
func {{.Receiver}}{{.Name}}(ctx context.Context, req <-chan *pb.{{.RequestType}}) (*pb.{{.ResponseType}}, error) {
	{{.Body}}
}
`
	streamOutSkeleton = `
func {{.Receiver}}{{.Name}}(ctx context.Context, req *pb.{{.RequestType}}, resp chan *pb.{{.ResponseType}}) error {
	{{.Body}}
}
`
)

var (
	tmpl              = template.Must(template.New("scaffold").Parse(scaffoldTemplate))
	UnaryMethod       = template.Must(template.New("unary").Parse(unarySkeleton))
	StreamInOutMethod = template.Must(template.New("stream-in-out").Parse(streamInOutSkeleton))
	StreamInMethod    = template.Must(template.New("stream-in").Parse(streamInSkeleton))
	StreamOutMethod   = template.Must(template.New("stream-out").Parse(streamOutSkeleton))
	UnaryError        = template.Must(template.New("unary-error").Parse(unaryErrorBody))
	StreamInOutError  = template.Must(template.New("stream-in-out-error").Parse(streamInOutErrorBody))
	StreamInError     = template.Must(template.New("stream-in-error").Parse(streamInErrorBody))
	StreamOutError    = template.Must(template.New("stream-out-error").Parse(streamOutErrorBody))
)
