package server

import (
	v1 "github.com/aisphereio/kernel-layout/api/todo/v1"
	"github.com/aisphereio/kernel-layout/internal/service"
	"github.com/aisphereio/kernel/serverx"
)

func TodoModules() []serverx.ServiceModule {
	return []serverx.ServiceModule{
		v1.TodoServiceKernelModule(),
	}
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
