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

package methods

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/mattmoor/korpc/pkg/install"
	"github.com/mattmoor/korpc/pkg/parameter"
	"github.com/mattmoor/korpc/pkg/protoplugin"
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

	generateCmds := []string{"package methods"}

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
						Name:            "methods",
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
	files := make([]string, 0, len(request.FileToGenerate))
	for _, file := range request.FileToGenerate {
		files = append(files, filepath.Join(stuff.NestingEscape(), file))
	}

	sdp, mdp := getSDPAndMDP(stuff, request)
	if sdp == nil || mdp == nil {
		return nil, fmt.Errorf("Unable to find %s.%s", stuff.Service, stuff.Method)
	}

	// We generate another level of //go:generate here instead of creating the
	// actual file because once development starts users are not going to want us
	// clobbering their functions every time they codegen.  By doing things this
	// way, they can use the following command to bootstrap:
	//    go generate ./pkg/methods/...
	// and as they add methods to their service, or add new services they can use:
	//    go generate ./pkg/methods/<lowercase service name>/<lowercase method name>
	// which will generate just the new method's scaffolding.
	binary, args := install.ProtoCCmd(
		".",
		install.KORPCPath,
		parameter.Stuff{
			Name:            "scaffold",
			Base:            stuff.Base,
			GenDir:          stuff.GenDir,
			MethodsDir:      stuff.MethodsDir,
			Domain:          stuff.Domain,
			Namespace:       stuff.Namespace,
			Service:         sdp.GetName(),
			Method:          mdp.GetName(),
			NestedDirectory: stuff.NestedDirectory,
		},
		files...,
	)

	generateCmds := []string{
		fmt.Sprintf("package %s", strings.ToLower(mdp.GetName())),
		fmt.Sprintf("//go:generate %s %s\n", binary, strings.Join(args, " ")),
	}

	mainName := "korpc.go"
	mainContent := strings.Join(generateCmds, "\n")

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

func init() {
	protoplugin.Register("methods", &plugin{})
}
