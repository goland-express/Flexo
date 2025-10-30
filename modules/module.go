package modules

import "github.com/goland-express/flexo/registry"

type Module interface {
	Name() string
	Register(r *registry.Registry)
}
