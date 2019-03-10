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

package config

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/mattmoor/korpc/pkg/install"
	"github.com/mattmoor/korpc/pkg/naming"
	"github.com/mattmoor/korpc/pkg/parameter"
	"github.com/mattmoor/korpc/pkg/protoplugin"

	korpc "github.com/mattmoor/korpc/include"
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

	generateCmds := []string{"package config"}

	var resp plugin_go.CodeGeneratorResponse
	for _, fd := range request.ProtoFile {
		if _, ok := codegen[fd.GetName()]; !ok {
			continue
		}
		for _, sdp := range fd.Service {
			for _, mdp := range sdp.Method {
				binary, args := install.ProtoCCmd(
					".",
					install.KORPCPath,
					parameter.Stuff{
						Name:            "config",
						Base:            stuff.Base,
						GenDir:          stuff.GenDir,
						MethodsDir:      stuff.MethodsDir,
						Service:         sdp.GetName(),
						Method:          mdp.GetName(),
						NestedDirectory: stuff.NestedDirectory,
					},
					files...,
				)

				generateCmds = append(generateCmds,
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

var tmpl = template.Must(template.New("service").Parse(serviceTemplate))

func (p *plugin) doMethod(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	sdp, mdp := getSDPAndMDP(stuff, request)
	if sdp == nil || mdp == nil {
		return nil, fmt.Errorf("Unable to find %s.%s", stuff.Service, stuff.Method)
	}

	opt := &options{
		Name:      naming.Service(sdp, mdp),
		Namespace: "default",
		// {base}/gen/entrypoint/{service}/{method}
		GatewayPath: filepath.Join(stuff.Base, stuff.GenDir, "entrypoint",
			strings.ToLower(sdp.GetName()), strings.ToLower(mdp.GetName())),
	}

	addr, err := proto.GetExtension(mdp.Options, korpc.E_Options)
	if err == nil {
		opt.Options = *(addr.(*korpc.Options))
	}

	mainName := naming.Service(sdp, mdp) + ".yaml"
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

func getSDPAndMDP(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*descriptor.ServiceDescriptorProto, *descriptor.MethodDescriptorProto) {
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
					return sdp, mdp
				}
			}
		}
	}
	return nil, nil
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
	protoplugin.Register("config", &plugin{})
}
