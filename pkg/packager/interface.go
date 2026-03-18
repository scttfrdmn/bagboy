// Package packager defines the Packager interface and Registry used by bagboy
// to create software packages for various platforms and package managers.
package packager

import (
	"context"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Packager is the interface implemented by every platform packager.
// It creates distribution artefacts (e.g. a Homebrew formula, a Debian package)
// from a bagboy Config.
type Packager interface {
	// Pack generates the package artefacts for the given config and returns
	// the path to the primary output file or directory.
	Pack(ctx context.Context, cfg *config.Config) (string, error)

	// Name returns the unique lower-case identifier for this packager
	// (e.g. "brew", "deb", "docker").
	Name() string

	// Validate checks whether cfg contains the fields required by this packager.
	// It returns a non-nil error when the config is insufficient.
	Validate(cfg *config.Config) error
}

// Registry is a collection of named Packagers.
type Registry struct {
	packagers map[string]Packager
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		packagers: make(map[string]Packager),
	}
}

// Register adds p to the registry, keyed by p.Name().
// Any existing packager with the same name is replaced.
func (r *Registry) Register(p Packager) {
	r.packagers[p.Name()] = p
}

// Get returns the Packager registered under name, and whether it exists.
func (r *Registry) Get(name string) (Packager, bool) {
	p, ok := r.packagers[name]
	return p, ok
}

// List returns the names of all registered packagers in unspecified order.
func (r *Registry) List() []string {
	var names []string
	for name := range r.packagers {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered packagers.
func (r *Registry) Count() int {
	return len(r.packagers)
}

// PackAll invokes every registered packager that passes validation and returns
// a map from packager name to output path. The first packager error aborts the
// run and returns that error.
func (r *Registry) PackAll(ctx context.Context, cfg *config.Config) (map[string]string, error) {
	results := make(map[string]string)

	for name, p := range r.packagers {
		if err := p.Validate(cfg); err != nil {
			continue // Skip packagers whose requirements are not met.
		}

		output, err := p.Pack(ctx, cfg)
		if err != nil {
			return nil, err
		}

		results[name] = output
	}

	return results, nil
}
