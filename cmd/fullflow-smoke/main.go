package main

import (
	"encoding/json"
	"fmt"

	"github.com/aisphereio/kernel-layout/internal/server"
)

type smokeResult struct {
	ModuleCount int      `json:"module_count"`
	Modules     []string `json:"modules"`
}

// fullflow-smoke now validates the current generated-service composition
// boundary. The previous command exercised the removed
// kernel/middleware/ratelimit compatibility package and could no longer build
// against the supported Kernel runtime API.
func main() {
	catalog := server.TodoCatalog()
	modules := catalog.Modules()
	if len(modules) != 1 {
		panic(fmt.Sprintf("unexpected Todo module count: %d", len(modules)))
	}

	result := smokeResult{
		ModuleCount: len(modules),
		Modules:     make([]string, 0, len(modules)),
	}
	for _, module := range modules {
		result.Modules = append(result.Modules, module.ModuleName())
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
