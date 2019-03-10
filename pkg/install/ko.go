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

package install

import (
	"os"
	"os/exec"
	"path/filepath"
)

var (
	KOPath = filepath.Join(Directory(), "bin", "ko")
)

func InstallKO() error {
	cmd := exec.Command("go", "get", "github.com/google/go-containerregistry/cmd/ko")

	// Pass through our environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOPATH="+Directory())

	// Pass through our stdfoo
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	// Run it.
	return cmd.Run()
}
