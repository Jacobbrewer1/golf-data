package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types -package api -templates ../../templates -o types.go -config ../../oapi-config.yaml ./routes.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate gorilla -package api -templates ../../templates -o server.go -config ../../oapi-config.yaml ./routes.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate client -package api -templates ../../templates -o client.go -config ../../oapi-config.yaml ./routes.yaml
