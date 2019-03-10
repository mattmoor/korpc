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

package gateway

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/mattmoor/korpc/pkg/naming"
	"github.com/mattmoor/korpc/pkg/parameter"
	"github.com/mattmoor/korpc/pkg/protoplugin"
)

type plugin struct {
}

var _ protoplugin.Interface = (*plugin)(nil)

var tmpl = template.Must(template.New("gateway").Parse(gatewayTemplate))

func (p *plugin) Do(stuff *parameter.Stuff, request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	codegen := make(map[string]struct{})
	for _, file := range request.FileToGenerate {
		codegen[file] = struct{}{}
	}

	var resp plugin_go.CodeGeneratorResponse
	opt := &options{
		Name:      "grpc-gateway",
		Namespace: stuff.Namespace,
		Domain:    stuff.Domain,
	}
	for _, fd := range request.ProtoFile {
		if _, ok := codegen[fd.GetName()]; !ok {
			continue
		}

		for _, sdp := range fd.Service {
			for _, mdp := range sdp.Method {
				opt.RoutingRules = append(opt.RoutingRules, routingRule{
					Path:        fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), sdp.GetName(), mdp.GetName()),
					ServiceName: naming.Service(sdp, mdp),
				})
			}
		}
	}

	// Based on the accumulated rules generate the dispatch yaml.
	mainName := "gateway.yaml"
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
	protoplugin.Register("gateway", &plugin{})
}
