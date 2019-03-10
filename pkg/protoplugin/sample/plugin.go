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

package sample

import (
	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/mattmoor/korpc/pkg/parameter"
	"github.com/mattmoor/korpc/pkg/protoplugin"
)

type plugin struct {
}

var _ protoplugin.Interface = (*plugin)(nil)

func (p *plugin) Do(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	codegen := make(map[string]struct{})
	for _, file := range request.FileToGenerate {
		codegen[file] = struct{}{}
	}

	var resp plugin_go.CodeGeneratorResponse
	// for _, fd := range request.ProtoFile {
	// 	if _, ok := codegen[fd.GetName()]; !ok {
	// 		continue
	// 	}
	// }

	mainName := "main.go"
	mainContent := "package main\n"

	resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
		Name:    &mainName,
		Content: &mainContent,
	})
	return &resp, nil
}

func init() {
	protoplugin.Register("Matt", &plugin{})
}
