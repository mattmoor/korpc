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
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/mattmoor/korpc/pkg/install"
	"github.com/mattmoor/korpc/pkg/parameter"
	"github.com/mattmoor/korpc/pkg/protoplugin"
	"github.com/mattmoor/korpc/pkg/protoplugin/scaffold"
)

type plugin struct {
}

var _ protoplugin.Interface = (*plugin)(nil)

func (p *plugin) Do(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	if stuff.Service != "" || stuff.Method != "" {
		return p.doMethod(stuff, request)
	}
	return p.doMeta(stuff, request)
}

func (p *plugin) doMeta(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	files := make([]string, 0, len(request.FileToGenerate))
	codegen := make(map[string]struct{})
	for _, file := range request.FileToGenerate {
		codegen[file] = struct{}{}
		files = append(files, filepath.Join(stuff.NestingEscape(), file))
	}

	generateCmds := []string{"package entrypoint"}

	var resp plugin_go.CodeGeneratorResponse
	for _, fd := range request.ProtoFile {
		if _, ok := codegen[fd.GetName()]; !ok {
			continue
		}
		for _, sdp := range fd.Service {
			for _, mdp := range sdp.Method {
				dir := filepath.Join(strings.ToLower(sdp.GetName()),
					strings.ToLower(mdp.GetName()))
				binary, args := install.ProtoCCmd(
					dir,
					install.KORPCPath,
					parameter.Stuff{
						Name:            "entrypoint",
						Base:            stuff.Base,
						GenDir:          stuff.GenDir,
						MethodsDir:      stuff.MethodsDir,
						Domain:          stuff.Domain,
						Namespace:       stuff.Namespace,
						Service:         sdp.GetName(),
						Method:          mdp.GetName(),
						NestedDirectory: filepath.Join(stuff.NestedDirectory, dir),
					},
					files...,
				)

				generateCmds = append(generateCmds,
					fmt.Sprintf("//go:generate mkdir -p %s", dir),
					fmt.Sprintf("//go:generate %s %s\n", binary, strings.Join(args, " ")))
			}
		}
	}

	mainName := "korpc.go"
	mainContent := strings.Join(generateCmds, "\n")

	resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
		Name:    &mainName,
		Content: &mainContent,
	})
	return &resp, nil
}

func (p *plugin) doMethod(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	codegen := make(map[string]struct{})
	for _, file := range request.FileToGenerate {
		codegen[file] = struct{}{}
	}

	opt := &options{
		ImplImportPath: filepath.Join(stuff.Base, stuff.MethodsDir,
			strings.ToLower(stuff.Service), strings.ToLower(stuff.Method)),
		Service: stuff.Service,
	}

	var resp plugin_go.CodeGeneratorResponse
	for _, fd := range request.ProtoFile {
		if _, ok := codegen[fd.GetName()]; !ok {
			continue
		}
		for _, sdp := range fd.Service {
			for _, mdp := range sdp.Method {
				if sdp.GetName() == stuff.Service && mdp.GetName() == stuff.Method {
					opt.ProtoImportPath = filepath.Join(stuff.Base, stuff.GenDir, "proto",
						// protoc-gen-go includes directory names
						filepath.Dir(fd.GetName()))
					var err error
					opt.Implementation, err = impl(sdp, mdp)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	mainName := "main.go"
	mainContent, err := execToString(tmpl, opt)
	if err != nil {
		return nil, err
	}

	resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
		Name:    &mainName,
		Content: &mainContent,
	})
	return &resp, nil
}

func impl(sdp *descriptor.ServiceDescriptorProto, mdp *descriptor.MethodDescriptorProto) (string, error) {
	opt := map[string]string{
		"Service":      sdp.GetName(),
		"Method":       mdp.GetName(),
		"Name":         mdp.GetName(),
		"RequestType":  extract(mdp.GetInputType()),
		"ResponseType": extract(mdp.GetOutputType()),
		"Receiver":     "(s *server) ",
	}
	switch {
	case mdp.GetServerStreaming() && mdp.GetClientStreaming():
		return execToString(streamInOutMethod, opt)
	case mdp.GetClientStreaming():
		return execToString(streamInMethod, opt)
	case mdp.GetServerStreaming():
		return execToString(streamOutMethod, opt)
	default:
		opt["Body"] = "return impl.Impl(ctx, req)"
		return execToString(scaffold.UnaryMethod, opt)
	}
}

// proto types come through as `.package.TypeName` so extract the last portion.
func extract(t string) string {
	parts := strings.Split(t, ".")
	return parts[len(parts)-1]
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
	protoplugin.Register("entrypoint", &plugin{})
}
