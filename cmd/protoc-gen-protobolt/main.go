package main

import (
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"

	"github.com/SpeedyCoder/protobolt/cmd/protoc-gen-protobolt/internal/repository"
)

func main() {
	pgs.Init(
		pgs.DebugEnv("DEBUG"),
	).RegisterModule(
		repository.NewRepositoryModule(),
	).RegisterPostProcessor(
		pgsgo.GoFmt(),
	).Render()
}
