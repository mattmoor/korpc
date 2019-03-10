# korpc

`korpc` (pronounced "corpse-ee") is an EXPERIMENTAL command-line tool that
aims to make GRPC API development as simple as FaaS / serverless development.

> TODO(mattmoor): Get a zombie gopher logo

In detail... `korpc` generates all of the scaffolding from a GRPC `.proto`
file so that you would literally just implement a Go `func` per RPC method.
These methods are then immediately deployable as N
[Knative](https://github.com/knative/serving) Service resources with a
single Istio VirtualService resource in front dispatching each RPC method to
a different Knative Service.

Amount of yaml written?  **Zero.**

### Installing `korpc`

`korpc` currently works on linux/amd64, to install it simply:

```shell
# Install korpc itself.
go get github.com/mattmoor/korpc/cmd/korpc

# Pull down korpc deps (e.g. protoc, ko, protoc-gen-go)
# These are installed alongside the korpc binary in a
# .korpc directory.
korpc install
```

You should only need to do this once per machine.

### How it works

`korpc` starts with a simple GRPC service definition:

```proto
syntax = "proto3";

package sample;

service SampleService {
  rpc Foo(FooRequest) returns (FooResponse) {}
  rpc Bar(BarRequest) returns (BarResponse) {}
}
```

Then in the root of your repository you simply add the following
to `doc.go` (really any `.go` file will do):

```go
# TODO(mattmoor): Support the domain/namespace here
//go:generate korpc generate --base=github.com/mattmoor/korpc-sample service.proto
```

Replace the argument to `--base=` with the path for your project, and then list your
own `.proto` files where you see `service.proto`.  You can now run:

```shell
# Run from the root diectory of your repo:
$ go generate .
2019/03/09 23:09:32 Generating proto...
2019/03/09 23:09:32 Generating entrypoint...
2019/03/09 23:09:32 Generating config...
2019/03/09 23:09:33 Generating gateway...
2019/03/09 23:09:33 Generating methods...
2019/03/09 23:09:33 korpc code-generation complete.
2019/03/09 23:09:33 To generate the skeleton for the RPC methods run:
  go generate ./pkg/methods/...
2019/03/09 23:09:33 To generate the skeleton for a single newly-added method run:
  go generate ./pkg/methods/<ServiceName>/<MethodName>
```

This is safe to run anytime your protos change, and produces usable GRPC client
code under `./gen/proto` (in addition to a whole bunch of other stuff).

By default `korpc` wants you to implement methods under `./pkg/methods`. For
example the method `Foo` in the example above would be implemented by a method
named `Impl` in `./pkg/methods/sampleservice/foo`. As the output of
`korpc generate` indicates, you can generate the scaffolding for these
individually, or in bulk (e.g. when bootstrapping a new project aka _now_) via:

```shell
go generate ./pkg/methods/...
```

> WARNING: Running this will overwrite all existing method definitions, so
> apply it carefully.

That's it.  You can now deploy a functioning GRPC service!

```shell
# TODO: Add support for `korpc deploy`
$ ko apply -f gen/config/
2019/03/09 23:30:37 Building github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/stream
2019/03/09 23:30:37 Building github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/unary
virtualservice.networking.istio.io/grpc-gateway unchanged
2019/03/09 23:30:40 Using base gcr.io/distroless/static:latest for github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/stream
2019/03/09 23:30:40 Using base gcr.io/distroless/static:latest for github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/unary
2019/03/09 23:30:40 Publishing us.gcr.io/convoy-adapter/unary-3127e44d6c8b83e970c0ffe9fa17f688:latest
2019/03/09 23:30:40 Publishing us.gcr.io/convoy-adapter/stream-e2352495a7144827a5c144e3140c7af7:latest
2019/03/09 23:30:41 existing blob: sha256:b5fe0529ab3e5680a2b47a447efc061ada2e15e9ffc9ec920458bd2d194e945d
2019/03/09 23:30:41 existing blob: sha256:4003b5b92ca98a8926d9112839f3f17e69f4ec4f995abb188a3ce3ccf93cd6d9
2019/03/09 23:30:41 existing blob: sha256:4003b5b92ca98a8926d9112839f3f17e69f4ec4f995abb188a3ce3ccf93cd6d9
2019/03/09 23:30:41 existing blob: sha256:aaced6115a76b945ac7aec28a3b2592b574fffbcd17abe4d92f05120289c429f
2019/03/09 23:30:41 existing blob: sha256:805a5e151e44f883809630c2c63bc3dd879fe3090ad7783da2a212904c36b422
2019/03/09 23:30:41 existing blob: sha256:4e1edcbff92b2fd48d837d1577dd69c4f00282a0ad675cbec0ec0ceec1384b65
2019/03/09 23:30:41 existing blob: sha256:4e1edcbff92b2fd48d837d1577dd69c4f00282a0ad675cbec0ec0ceec1384b65
2019/03/09 23:30:41 existing blob: sha256:758dbf4df7dca7256dabd8a6f49073a48109765d7e8fe08a95118b18ef9e1e2f
2019/03/09 23:30:42 us.gcr.io/convoy-adapter/unary-3127e44d6c8b83e970c0ffe9fa17f688:latest: digest: sha256:503667f2e934c2cb4c62502fec851811d0bd87dd47f14aff36d82996f4d74fd5 size: 750
2019/03/09 23:30:42 Published us.gcr.io/convoy-adapter/unary-3127e44d6c8b83e970c0ffe9fa17f688@sha256:503667f2e934c2cb4c62502fec851811d0bd87dd47f14aff36d82996f4d74fd5
2019/03/09 23:30:42 us.gcr.io/convoy-adapter/stream-e2352495a7144827a5c144e3140c7af7:latest: digest: sha256:07676525f15e40eefe2cb3f232938fa2662fd5ca1674cf8f88669b19b47f436a size: 750
2019/03/09 23:30:42 Published us.gcr.io/convoy-adapter/stream-e2352495a7144827a5c144e3140c7af7@sha256:07676525f15e40eefe2cb3f232938fa2662fd5ca1674cf8f88669b19b47f436a
service.serving.knative.dev/stream-sampleservice configured
service.serving.knative.dev/unary-sampleservice configured
```

[This example](https://github.com/mattmoor/korpc-sample) is deploying a
two-method GRPC service, which will become three Kubernetes resources:
1. A VirtualService that acts as an API Gateway and routes each GRPC method
  to the appropriate Service.
1. A Knative Service for the "Unary" method.
1. A Knative Service for the "Stream" method.

At this point, all this service's methods do is return errors about not
being implemented, but the starter scaffolding is in place. Open up the
generated stubs in `./pkg/methods/...` and start hacking!

> PRO TIP: as you are iterating, leave `ko apply -w -f ./gen/config` running
> it will watch and automatically rebuild/redeploy things whenever you make an
> edit. It will **not** however rerun code generation, so whenever you change
> your protos be sure to rerun `go generate .`


### Customizing the Knative Services

A key aspect of how `korpc` works is putting each GRPC method into a separate
Knative Service. The intent behind this is to enable each method to be
configured differently.  For example, a user may want their writes to go to
Services with larger memory allocations. Or they may be security conscious
and want different operations to run as distinct identities (e.g. reads
shouldn't write). To configure some of these options, you simply decorate
the RPC method as follows:

```proto
syntax = "proto3";

// Import the needed options for customizing things.
import "github.com/mattmoor/korpc/include/korpc.proto";

package sample;

service SampleService {
  rpc Foo(FooRequest) returns (FooResponse) {
    // Decorate the method with the options you want.
    option (korpc.options) = {
      service_account = "restricted-to-foo"
      container_concurrency: 20
      env: {
        name: "FOO"
        value: "BAR"
      }
      env: {
        name: "UGH"
        value: "GAZUNK"
      }
    }
  }
  ...
}

```

> This is the extent of what's currently supported, but adding more
> options is a simple exercise is plumbing.


### `.git{ignore,attributes}` recommendations

`korpc` generates a lot of files to accomplish its task. We recommend adding
`/**/korpc.go` to `.gitignore` because these files contain install-relative paths
and are not needed for building.

In addition, we would recommend adding the following lines to `.gitattributes`
so that Github code reviews will hide them until expanded:

```
/gen/** linguist-generated=true
/**/korpc.go linguist-generated=true
```


### TODO list
> TODO: don't hardcode domain.
> TODO: don't hardcode namespace.
> TODO: add `korpc deploy` using the installed `ko`.
