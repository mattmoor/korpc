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

package gateway

type options struct {
	Name         string
	Namespace    string
	Domain       string
	RoutingRules []routingRule
}

type routingRule struct {
	Path        string
	ServiceName string
}

const (
	gatewayTemplate = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: {{$.Name}}
  namespace: {{$.Namespace}}
spec:
  gateways:
  - knative-ingress-gateway.knative-serving.svc.cluster.local
  - mesh
  hosts:
  - {{$.Domain}}
  http:
{{range $val := .RoutingRules}}
  - match:
    - uri:
        exact: {{$val.Path}}
    rewrite:
      authority: {{$val.ServiceName}}.{{$.Namespace}}.svc.cluster.local
    route:
      - destination:
          host: istio-ingressgateway.istio-system.svc.cluster.local
          port:
            number: 80
        weight: 100
{{end}}
`
)
