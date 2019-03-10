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

package protoplugin

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/mattmoor/korpc/pkg/parameter"
)

// Interface defines how we interact with proto plugins.
type Interface interface {
	Do(*parameter.Stuff, *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error)
}

var (
	m       sync.Mutex
	plugins = map[string]Interface{}
)

func Register(name string, thing Interface) {
	m.Lock()
	defer m.Unlock()
	if _, ok := plugins[name]; ok {
		log.Fatalf("Duplicate plugin named %q registered", name)
	}
	plugins[name] = thing
}

func Get(name string) Interface {
	m.Lock()
	defer m.Unlock()
	return plugins[name]
}

func dispatch(request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	stuff, err := parameter.From(request.GetParameter())
	if err != nil {
		return nil, err
	}
	plugin := Get(stuff.Name)
	if plugin == nil {
		return nil, fmt.Errorf("Unrecognized plugin %q", stuff.Name)
	}
	return plugin.Do(stuff, request)
}

func Run() error {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	var req plugin_go.CodeGeneratorRequest
	if err := req.Unmarshal(bytes); err != nil {
		return err
	}

	resp, err := dispatch(&req)
	if err != nil {
		// Return plugin errors as an error response instead of a
		// non-zero exit code.
		s := err.Error()
		resp = &plugin_go.CodeGeneratorResponse{
			Error: &s,
		}
	}

	result, err := resp.Marshal(nil, true)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(result)
	return err
}
