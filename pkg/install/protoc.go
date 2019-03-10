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
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattmoor/korpc/pkg/parameter"
)

const (
	version      = "3.7.0"
	platform     = "linux"
	architecture = "x86_64"
)

var (
	downloadURL = fmt.Sprintf("https://github.com/protocolbuffers/protobuf/releases/download/v%[1]s/protoc-%[1]s-%[2]s-%[3]s.zip", version, platform, architecture)

	protocInstallDir = filepath.Join(Directory(), fmt.Sprintf("protoc-%s", version))

	ProtoCBinary  = filepath.Join(protocInstallDir, "bin", "protoc")
	ProtoCInclude = filepath.Join(protocInstallDir, "include")
)

func toroot(out string, param parameter.Stuff) string {
	invert := func(p string) string {
		return (&parameter.Stuff{NestedDirectory: p}).NestingEscape()
	}

	return invert(filepath.Join(param.NestedDirectory, invert(out)))
}

func ProtoCCmd(out string, plugin string, param parameter.Stuff, protos ...string) (string, []string) {
	args := []string{
		"-I" + ProtoCInclude,
		"-I" + KORPCInclude,
		"-I" + toroot(out, param),
		"--plugin=protoc-gen-" + param.Name + "=" + plugin,
	}

	switch plugin {
	case KORPCPath:
		args = append(args,
			"--"+param.Name+"_out="+out,
			"--"+param.Name+"_opt="+param.MustEncode(),
		)
	case ProtoCGenGoPath:
		args = append(args,
			"--"+param.Name+"_out=plugins=grpc:"+out,
		)
	}

	args = append(args, protos...)
	return ProtoCBinary, args
}

func RunProtoC(out string, plugin string, param parameter.Stuff, protos ...string) error {
	binary, args := ProtoCCmd(out, plugin, param, protos...)

	cmd := exec.Command(binary, args...)

	// Pass through our environment
	cmd.Env = os.Environ()

	// Pass through our stdfoo
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	// Run it.
	return cmd.Run()
}

// This is based on https://golangcode.com/unzip-files-in-go/
func InstallProtoC() error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status fetching protoc release: %d", resp.StatusCode)
	}

	// It sucks that we need an io.ReaderAt, which forces us to read things into memory.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	dest := protocInstallDir
	log.Printf("Installing to: %s", dest)
	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
