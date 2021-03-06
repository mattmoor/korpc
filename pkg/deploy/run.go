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

package deploy

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mattmoor/korpc/pkg/install"
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

func koapply() error {
	cmd := exec.Command(install.KOPath, "apply", "-f", filepath.Join(gen, "config"))

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
	if err := gogenerate("."); err != nil {
		log.Fatalf("Error generating API: %v", err)
	}
	if err := koapply(); err != nil {
		log.Fatalf("Error deploying API: %v", err)
	}
}
