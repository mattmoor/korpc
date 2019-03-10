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

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mattmoor/korpc/pkg/delete"
	"github.com/mattmoor/korpc/pkg/deploy"
	"github.com/mattmoor/korpc/pkg/generate"
	"github.com/mattmoor/korpc/pkg/install"
	"github.com/mattmoor/korpc/pkg/protoplugin"

	// The protoc plugins that we have enabled.
	_ "github.com/mattmoor/korpc/pkg/protoplugin/config"
	_ "github.com/mattmoor/korpc/pkg/protoplugin/entrypoint"
	_ "github.com/mattmoor/korpc/pkg/protoplugin/gateway"
	_ "github.com/mattmoor/korpc/pkg/protoplugin/methods"
	_ "github.com/mattmoor/korpc/pkg/protoplugin/scaffold"
	// _ "github.com/mattmoor/korpc/pkg/protoplugin/sample"
)

func main() {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "korpc",
		Short: "A tool for GRPC APIaaS development in Go with ko",
		Run: func(cmd *cobra.Command, args []string) {
			// Based on this:
			// https://rosettacode.org/wiki/Check_output_device_is_a_terminal#Go
			if terminal.IsTerminal(int(os.Stdin.Fd())) {
				cmd.Help()
				return
			}

			// When someone is streaming input (not run on a terminal) act as though
			// we're being invoked by protoc.
			if err := protoplugin.Run(); err != nil {
				cmd.Help()
				log.Fatalf("Error executing proto plugin: %v", err)
			}
		},
	}

	cmds.AddCommand(deploy.Command)
	cmds.AddCommand(delete.Command)
	cmds.AddCommand(generate.Command)
	cmds.AddCommand(install.Command)

	if err := cmds.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
