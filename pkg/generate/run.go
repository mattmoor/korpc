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

package generate

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mattmoor/korpc/pkg/install"
	"github.com/mattmoor/korpc/pkg/parameter"
)

func gogenerate(pkg string) error {
	cmd := exec.Command("go", "generate", pkg)

	// Pass through our environment
	cmd.Env = os.Environ()

	// Pass through our stdfoo
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	// Run it.
	return cmd.Run()
}

func run(cmd *cobra.Command, args []string) {
	invocations := []struct {
		PluginPath string
		Params     parameter.Stuff
		Generate   bool
	}{{
		PluginPath: install.ProtoCGenGoPath,
		Params: parameter.Stuff{
			Name:            "proto",
			Base:            base,
			GenDir:          gen,
			MethodsDir:      methods,
			NestedDirectory: filepath.Join(gen, "proto"),
		},
	}, {
		PluginPath: install.KORPCPath,
		Params: parameter.Stuff{
			Name:            "entrypoint",
			Base:            base,
			GenDir:          gen,
			MethodsDir:      methods,
			NestedDirectory: filepath.Join(gen, "entrypoint"),
		},
		Generate: true,
	}, {
		PluginPath: install.KORPCPath,
		Params: parameter.Stuff{
			Name:            "config",
			Base:            base,
			GenDir:          gen,
			MethodsDir:      methods,
			NestedDirectory: filepath.Join(gen, "config"),
		},
		Generate: true,
	}, {
		PluginPath: install.KORPCPath,
		Params: parameter.Stuff{
			Name:       "gateway",
			Base:       base,
			GenDir:     gen,
			MethodsDir: methods,
			// Put the gateway into config.
			NestedDirectory: filepath.Join(gen, "config"),
		},
		Generate: false,
	}, {
		PluginPath: install.KORPCPath,
		Params: parameter.Stuff{
			Name:            "methods",
			Base:            base,
			GenDir:          gen,
			MethodsDir:      methods,
			NestedDirectory: filepath.Join(methods),
		},
		Generate: true,
	}}

	for _, inv := range invocations {
		log.Printf("Generating %s...", inv.Params.Name)
		if err := os.MkdirAll(inv.Params.NestedDirectory, 0777); err != nil {
			log.Fatalf("Error creating output directory %q: %v", inv.Params.NestedDirectory, err)
		}
		if err := install.RunProtoC(inv.Params.NestedDirectory, inv.PluginPath, inv.Params, args...); err != nil {
			log.Fatalf("Error running protoc on %q: %v", inv.Params.Name, err)
		}
		if !inv.Generate {
			continue
		}
		if err := gogenerate("./" + inv.Params.NestedDirectory); err != nil {
			log.Fatalf("Error running go generate on %q: %v", inv.Params.NestedDirectory, err)
		}
	}

	log.Print("korpc code-generation complete.")
	log.Print("To generate the skeleton for the RPC methods run:\n  go generate ./pkg/methods/...")
	log.Print("To generate the skeleton for a single newly-added method run:\n  go generate ./pkg/methods/<ServiceName>/<MethodName>")

}
