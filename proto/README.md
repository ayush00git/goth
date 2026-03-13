we write auth.proto and running an external plugin protoc generates us two files using its stdin (CodeGeneratorRequest) and stdout(CodeGeneratorResponse) protocol.

we can run this command using the following command:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/auth.proto
```

this command generates two files =>
1. `auth.pb.go` - contains the model code, and is generated through the message defined in the `.proto` file.
2. `auth_grpc.pb.go` - contains the client and server stub implementation with the gPRC wiring. Generated through the service blocks in the `.proto` file.

client stub defines how the client would call the methods
server stub defines how the server would implement the methods
server stub also contains the mustEmbedUnimplemented method to all server interface methods to handle forward compatibility.

also contains the registrar that registers our server with the grpc server
contains the handlers which maps the incoming request from the client to the specific function.