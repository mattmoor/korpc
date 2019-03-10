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

package parameter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
)

type Stuff struct {
	Name            string `json:"name"`
	Base            string `json:"base,omitempty"`
	GenDir          string `json:"gen_dir,omitempty"`
	MethodsDir      string `json:"methods_dir,omitempty"`
	Service         string `json:"service,omitempty"`
	Method          string `json:"method,omitempty"`
	NestedDirectory string `json:"nested_directory,omitempty"`
}

var _ fmt.Stringer = (*Stuff)(nil)

// String implements fmt.Stringer
func (s *Stuff) String() string {
	return fmt.Sprintf("%s (%s.%s)", s.Name, s.Service, s.Method)
}

func (s *Stuff) NestingEscape() string {
	if s.NestedDirectory == "" {
		return ""
	}
	re := regexp.MustCompile("[^/.]+")
	return string(re.ReplaceAll([]byte(s.NestedDirectory), []byte("..")))
}

func (s *Stuff) Encode() (string, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func (s *Stuff) MustEncode() string {
	encoded, err := s.Encode()
	if err != nil {
		log.Fatalf("Failure during MustEncode for %q: %v", s.Name, err)
	}
	return encoded
}

func From(encoded string) (*Stuff, error) {
	stuff := &Stuff{}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(decoded, stuff); err != nil {
		return nil, err
	}
	return stuff, nil
}
