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
	"github.com/spf13/cobra"
)

var (
	base      string
	gen       string
	methods   string
	domain    string
	namespace string

	Command = &cobra.Command{
		Use:   "generate",
		Short: "Generate the yaml and entrypoints for the API.",
		Run:   run,
		Args:  cobra.MinimumNArgs(1),
	}
)

func init() {
	Command.Flags().StringVarP(&base, "base", "B", "",
		"The base import path for packages in this repository.")

	Command.Flags().StringVarP(&gen, "gen", "G", "./gen",
		"The directory under which to put generated code and configuration.")

	Command.Flags().StringVarP(&methods, "methods", "M", "./pkg/methods",
		"The directory under which to find RPC method implementations.")

	Command.Flags().StringVarP(&namespace, "namespace", "n", "default",
		"The namespace into which we should deploy things.")

	Command.Flags().StringVarP(&domain, "domain", "D", "",
		"The domain on which Istio will serve the resulting API.")
}
