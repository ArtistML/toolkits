# toolkits

## handbook for gen-grpc
1. create generate.go in service dir
2. add go generate line in generate.go files, e.g.: [generate.go](./test/grpc/generate.go)

## handbook for gen-proto
1. create common_resources.proto for each service proto, e.g.: [v1/common_resources.proto](./test/proto/echo/v1/common_resources.proto) [v2/common_resources.proto](./test/proto/echo/v2/common_resources.proto)
2. create entity proto which you want to generate CRUD api for, e.g.: [v1/greeting.proto](./test/proto/echo/v1/greeting.proto) [v2/greeting.proto](./test/proto/echo/v2/greeting.proto)
3. save your custom messages into types.proto, save your enum messages into enums.proto, e.g.: [v1/types.proto](./test/proto/echo/v1/types.proto) [v2/enums.proto](./test/proto/echo/v2/enums.proto)
4. create generate.go in proto project dir, e.g.: [generate.go](./test/proto/generate.go)
5. add filters: `-f echo/v2 -f echo/v1` means generate CRUD api for echo/v2 and echo/v1 only.

## handbook for gen-pypkg
1. `git clone https://github.com/ArtistML/toolkits.git`
2. `cd tollkits`
3. create local/generate.go file to generate python package, e.g.: [generate.go](./test/pypkg/generate.go)
