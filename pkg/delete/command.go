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

package delete

import (
	"github.com/spf13/cobra"
)

var (
	gen string

	Command = &cobra.Command{
		Use:   "delete",
		Short: "Delete the API from the current kubectl context.",
		Run:   run,
	}
)

func init() {
	Command.Flags().StringVarP(&gen, "gen", "G", "./gen",
		"The directory containing the generated code and configuration.")
}
