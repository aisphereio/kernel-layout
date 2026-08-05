package server

import (
	"fmt"

	v1 "github.com/aisphereio/kernel-layout/api/todo/v1"
	"github.com/aisphereio/kernel-layout/internal/service"
	"github.com/aisphereio/kernel/gatewayx"
	"github.com/aisphereio/kernel/serverx"
	kgrpc "github.com/aisphereio/kernel/transportx/grpc"
	khttp "github.com/aisphereio/kernel/transportx/http"
)

// TodoModules is the single declaration point for the generated Todo service
// contract. The registration hooks are supplied here so the current committed
// generated module remains usable until the next `make api` regeneration with
// the latest Kernel generator.
func TodoModules() []serverx.ServiceModule {
	module := v1.TodoServiceKernelModule()
	module.RegisterGRPC = func(s *kgrpc.Server, impl any) error {
		typed, ok := impl.(v1.TodoServiceServer)
		if !ok {
			return fmt.Errorf("serverx: service implementation for todo.v1.TodoService must implement TodoServiceServer, got %T", impl)
		}
		v1.RegisterTodoServiceServer(s, typed)
		return nil
	}
	module.RegisterHTTP = func(s *khttp.Server, impl any) error {
		typed, ok := impl.(v1.TodoServiceHTTPServer)
		if !ok {
			return fmt.Errorf("serverx: service implementation for todo.v1.TodoService must implement TodoServiceHTTPServer, got %T", impl)
		}
		v1.RegisterTodoServiceHTTPServer(s, typed)
		return nil
	}
	module.RegisterGatewayInvokers = func(registry *gatewayx.InvokerRegistry, client any) error {
		typed, ok := client.(v1.TodoServiceClient)
		if !ok {
			return fmt.Errorf("serverx: gateway client for todo.v1.TodoService must implement TodoServiceClient, got %T", client)
		}
		return v1.RegisterTodoServiceGatewayInvokers(registry, typed)
	}
	return []serverx.ServiceModule{module}
}

func TodoCatalog() serverx.ServiceCatalog {
	return serverx.MustServiceCatalog(TodoModules()...)
}

func TodoBindings(todo *service.TodoService) []serverx.ServiceBinding {
	modules := TodoModules()
	return []serverx.ServiceBinding{
		{Module: modules[0], Implementation: todo},
	}
}
