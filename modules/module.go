package modules

import "flexo/registry"

type Module interface {
	Name() string
	Register(r *registry.Registry)
}
