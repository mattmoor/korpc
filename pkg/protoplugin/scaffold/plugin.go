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
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/mattmoor/korpc/pkg/parameter"
	"github.com/mattmoor/korpc/pkg/protoplugin"
)

type plugin struct {
}

var _ protoplugin.Interface = (*plugin)(nil)

func (p *plugin) Do(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	fd, sdp, mdp := getDescriptors(stuff, request)
	if fd == nil || sdp == nil || mdp == nil {
		return nil, fmt.Errorf("Unable to find %s.%s", stuff.Service, stuff.Method)
	}

	body, err := unimpl(sdp, mdp)
	if err != nil {
		return nil, err
	}

	opt := &options{
		Package: strings.ToLower(mdp.GetName()),
		ProtoImportPath: filepath.Join(stuff.Base, stuff.GenDir, "proto",
			// protoc-gen-go includes directory names
			filepath.Dir(fd.GetName())),
		Body: body,
	}

	mainName := "main.go"
	mainContent, err := execToString(tmpl, opt)
	if err != nil {
		return nil, err
	}

	var resp plugin_go.CodeGeneratorResponse
	resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
		Name:    &mainName,
		Content: &mainContent,
	})
	return &resp, nil
}

func unimpl(sdp *descriptor.ServiceDescriptorProto, mdp *descriptor.MethodDescriptorProto) (string, error) {
	t, et := UnaryMethod, UnaryError
	switch {
	case mdp.GetServerStreaming() && mdp.GetClientStreaming():
		t, et = StreamInOutMethod, StreamInOutError
	case mdp.GetClientStreaming():
		t, et = StreamInMethod, StreamInError
	case mdp.GetServerStreaming():
		t, et = StreamOutMethod, StreamOutError
	}

	body, err := execToString(et, fmt.Sprintf("You need to implement %s.%s!!!",
		sdp.GetName(), mdp.GetName()))
	if err != nil {
		return "", err
	}

	opt := map[string]string{
		"Service":      sdp.GetName(),
		"Method":       mdp.GetName(),
		"RequestType":  extract(mdp.GetInputType()),
		"ResponseType": extract(mdp.GetOutputType()),
		// The package name is the RPC method, so in this context we simply name
		// the method "Impl" to avoid a stutter.  The method is also receiverless.
		"Name":     "Impl",
		"Receiver": "",
		"Body":     body,
	}
	return execToString(t, opt)
}

// proto types come through as `.package.TypeName` so extract the last portion.
func extract(t string) string {
	parts := strings.Split(t, ".")
	return parts[len(parts)-1]
}

func getDescriptors(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*descriptor.FileDescriptorProto, *descriptor.ServiceDescriptorProto, *descriptor.MethodDescriptorProto) {
	codegen := make(map[string]struct{})
	for _, file := range request.FileToGenerate {
		codegen[file] = struct{}{}
	}
	for _, fd := range request.ProtoFile {
		if _, ok := codegen[fd.GetName()]; !ok {
			continue
		}
		for _, sdp := range fd.Service {
			for _, mdp := range sdp.Method {
				if sdp.GetName() == stuff.Service && mdp.GetName() == stuff.Method {
					return fd, sdp, mdp
				}
			}
		}
	}
	return nil, nil, nil
}

// execute a template to produce a string.
func execToString(t *template.Template, opt interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := t.Execute(buf, opt)
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}

func init() {
	protoplugin.Register("scaffold", &plugin{})
}
