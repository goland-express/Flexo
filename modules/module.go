package modules

import "github.com/goland-express/Flexo/registry"

type Module interface {
	Name() string
	Register(r *registry.Registry)
}
