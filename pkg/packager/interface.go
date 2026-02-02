package packager

import (
	"context"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager interface {
	Pack(ctx context.Context, cfg *config.Config) (string, error)
	Name() string
	Validate(cfg *config.Config) error
}

type Registry struct {
	packagers map[string]Packager
}

func NewRegistry() *Registry {
	return &Registry{
		packagers: make(map[string]Packager),
	}
}

func (r *Registry) Register(p Packager) {
	r.packagers[p.Name()] = p
}

func (r *Registry) Get(name string) (Packager, bool) {
	p, ok := r.packagers[name]
	return p, ok
}

func (r *Registry) List() []string {
	var names []string
	for name := range r.packagers {
		names = append(names, name)
	}
	return names
}

func (r *Registry) Count() int {
	return len(r.packagers)
}

func (r *Registry) PackAll(ctx context.Context, cfg *config.Config) (map[string]string, error) {
	results := make(map[string]string)

	for name, packager := range r.packagers {
		if err := packager.Validate(cfg); err != nil {
			continue // Skip packagers that can't handle this config
		}

		output, err := packager.Pack(ctx, cfg)
		if err != nil {
			return nil, err
		}

		results[name] = output
	}

	return results, nil
}
