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

package config

import (
	korpc "github.com/mattmoor/korpc/include"
)

type options struct {
	Name        string
	Namespace   string
	GatewayPath string
	MethodLower string
	Options     korpc.Options
}

const (
	serviceTemplate = `apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: {{$.Name}}
  namespace: {{$.Namespace}}
spec:
  template:
    spec:
      {{if ne "" $.Options.ServiceAccount}}serviceAccountName: {{$.Options.ServiceAccount}}{{end}}
      {{if ne 0 $.Options.ContainerConcurrency}}containerConcurrency: {{$.Options.ContainerConcurrency}}{{end}}
      {{if ne 0 $.Options.TimeoutSeconds}}timeoutSeconds: {{$.Options.TimeoutSeconds}}{{end}}
      containers:
      - image: {{.GatewayPath}}
        ports:
        - name: h2c
          containerPort: 8080
        readinessProbe:
          exec:
            command: ["/ko-app/{{$.MethodLower}}", "probe"]
        livenessProbe:
          exec:
            command: ["/ko-app/{{$.MethodLower}}", "probe"]
        env:{{range $val := $.Options.Env}}
        - name: {{$val.Name}}
          value: {{$val.Value}}{{end}}
        resources:
          limits:{{range $key, $value := $.Options.GetResources.GetLimits}}
            {{$key}}: {{$value}}{{end}}
          requests:{{range $key, $value := $.Options.GetResources.GetRequests}}
            {{$key}}: {{$value}}{{end}}
`
)
