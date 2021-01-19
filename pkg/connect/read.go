package connect

import (
	"context"

	"github.com/containerd/containerd/remotes"
	"github.com/google/go-containerregistry/pkg/v1"
)

// Get get a reference to an image that can be used.
func Get(ref string, resolver remotes.Resolver) (*Descriptor, error) {
	ctx := context.Background()
	name, desc, err := resolver.Resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	v1Desc, err := v1.ParseDescriptor(desc)
	if err != nil {
		return nil, err
	}
	return &Descriptor{
		name:     name,
		resolver: resolver,
		desc:     v1Desc,
	}, nil
}
