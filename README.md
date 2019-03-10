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
//go:generate korpc generate --base=github.com/mattmoor/korpc-sample --domain=mattmoor.io service.proto
```

Replace the argument to `--base=` with the path for your project, `--domain=`
with the domain on which to serve, and then list your own `.proto` files where
you see `service.proto`.  You can now run:

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

> NOTE: After running this for the first time, you will have to ensure that your
> working tree has all of the needed dependencies e.g. via `dep ensure` or
> `dep init`.


> WARNING: Running this will overwrite all existing method definitions, so
> apply it carefully.

That's it.  You can now deploy a functioning GRPC service!

```shell
# If codegen isn't needed, consider using: ko apply -f ./gen/config
$ korpc deploy
2019/03/10 00:30:37 Installing to: /usr/local/google/home/mattmoor/go/bin/.korpc/protoc-3.7.0
2019/03/10 00:30:37 Generating proto...
2019/03/10 00:30:37 Generating entrypoint...
2019/03/10 00:30:37 Generating config...
2019/03/10 00:30:37 Generating gateway...
2019/03/10 00:30:37 Generating methods...
2019/03/10 00:30:38 korpc code-generation complete.
2019/03/10 00:30:38 To generate the skeleton for the RPC methods run:
  go generate ./pkg/methods/...
2019/03/10 00:30:38 To generate the skeleton for a single newly-added method run:
  go generate ./pkg/methods/<ServiceName>/<MethodName>
2019/03/10 00:30:38 Building github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/stream
2019/03/10 00:30:38 Building github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/unary
virtualservice.networking.istio.io/grpc-gateway unchanged
2019/03/10 00:30:40 Using base gcr.io/distroless/static:latest for github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/stream
2019/03/10 00:30:40 Using base gcr.io/distroless/static:latest for github.com/mattmoor/korpc-sample/gen/entrypoint/sampleservice/unary
2019/03/10 00:30:41 Publishing us.gcr.io/convoy-adapter/unary-3127e44d6c8b83e970c0ffe9fa17f688:latest
2019/03/10 00:30:41 Publishing us.gcr.io/convoy-adapter/stream-e2352495a7144827a5c144e3140c7af7:latest
2019/03/10 00:30:41 existing blob: sha256:4e1edcbff92b2fd48d837d1577dd69c4f00282a0ad675cbec0ec0ceec1384b65
2019/03/10 00:30:41 existing blob: sha256:4003b5b92ca98a8926d9112839f3f17e69f4ec4f995abb188a3ce3ccf93cd6d9
2019/03/10 00:30:41 existing blob: sha256:b28ef633cdf46402dbfebd680ff800cb5064741c23490ec45848f117e8fd5b65
2019/03/10 00:30:41 existing blob: sha256:59d0f749366a464c46bb3eaa8b363018e05a1ef8fa2c72add3211921b10335d8
2019/03/10 00:30:41 existing blob: sha256:4b9972d9e8791f5b95ea9a27a2e6325a72909da9ac7c681d4cde1784e5963a5f
2019/03/10 00:30:41 existing blob: sha256:3b6d39948647812b4dac05aab3101b2409fe8c75ab5e11b5812b6fb0dcb515e3
2019/03/10 00:30:41 existing blob: sha256:4e1edcbff92b2fd48d837d1577dd69c4f00282a0ad675cbec0ec0ceec1384b65
2019/03/10 00:30:41 existing blob: sha256:4003b5b92ca98a8926d9112839f3f17e69f4ec4f995abb188a3ce3ccf93cd6d9
2019/03/10 00:30:42 us.gcr.io/convoy-adapter/unary-3127e44d6c8b83e970c0ffe9fa17f688:latest: digest: sha256:e94cfa78617b78a0024a13e7f4b816c1fea97fc91a21a856341d84f764c5faa9 size: 750
2019/03/10 00:30:42 Published us.gcr.io/convoy-adapter/unary-3127e44d6c8b83e970c0ffe9fa17f688@sha256:e94cfa78617b78a0024a13e7f4b816c1fea97fc91a21a856341d84f764c5faa9
2019/03/10 00:30:42 us.gcr.io/convoy-adapter/stream-e2352495a7144827a5c144e3140c7af7:latest: digest: sha256:a9b9434d16e2ea7015c0f41bdab3a48f8d547d5c41ff1eb664a4666b1fa32997 size: 750
2019/03/10 00:30:42 Published us.gcr.io/convoy-adapter/stream-e2352495a7144827a5c144e3140c7af7@sha256:a9b9434d16e2ea7015c0f41bdab3a48f8d547d5c41ff1eb664a4666b1fa32997
service.serving.knative.dev/stream-sampleservice unchanged
service.serving.knative.dev/unary-sampleservice unchanged
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


### Cleaning up deployed APIs.

Similar to `korpc deploy` you can simply `korpc delete` to tear down the
deployed API.

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
