package remote

import (
	"context"
	"net/http"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/remotes"
	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Resolver implements github.com/containerd/containerd/remotes.Resolver
type Resolver struct {
	options    []Option
	acceptable []types.MediaType
}

func NewResolver(opts ...Option) Resolver {
	acceptable := []types.MediaType{
		// Just to look at them.
		types.DockerManifestSchema1,
		types.DockerManifestSchema1Signed,
	}
	acceptable = append(acceptable, acceptableImageMediaTypes...)
	acceptable = append(acceptable, acceptableIndexMediaTypes...)
	return Resolver{
		options:    opts,
		acceptable: acceptable,
	}
}

func (r Resolver) Resolve(ctx context.Context, ref string) (name string, desc ocispec.Descriptor, err error) {
	nref, err := namepkg.ParseReference(ref)
	if err != nil {
		return name, desc, err
	}
	o, err := makeOptions(nref.Context(), r.options...)
	if err != nil {
		return name, desc, err
	}
	f, err := makeFetcher(nref, o)
	if err != nil {
		return name, desc, err
	}
	b, v1desc, err := f.fetchManifest(nref, r.acceptable)
	if err != nil {
		return name, desc, err
	}
	desc, err = v1desc.Ocispec()
	return ref, desc, err
}

func (r Resolver) Fetcher(ctx context.Context, ref string) (remotes.Fetcher, error) {
	nref, err := namepkg.ParseReference(ref)
	if err != nil {
		return nil, err
	}
	o, err := makeOptions(nref.Context(), r.options...)
	if err != nil {
		return nil, err
	}
	f, err := makeFetcher(nref, o)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r Resolver) Pusher(ctx context.Context, ref string) (remotes.Pusher, error) {
	nref, err := namepkg.ParseReference(ref)
	if err != nil {
		return nil, err
	}
	o, err := makeOptions(nref.Context(), r.options...)
	if err != nil {
		return nil, err
	}
	return pusher{
		ref:     ref,
		options: o,
	}, nil
}

type pusher struct {
	ref     string
	options *options
}

// Push get a writer to push a descriptor
func (p pusher) Push(ctx context.Context, d ocispec.Descriptor) (content.Writer, error) {
	ref, err := namepkg.ParseReference(p.ref)
	if err != nil {
		return nil, err
	}
	scopes := []string{ref.Scope(transport.PushScope)}
	tr, err := transport.NewWithContext(p.options.context, ref.Context().Registry, p.options.auth, p.options.transport, scopes)
	if err != nil {
		return nil, err
	}
	writer := &writer{
		repo:    ref.Context(),
		client:  &http.Client{Transport: tr},
		context: p.options.context,
	}

	location, _, err := writer.initiateUpload("", d.Digest.String())
	if err != nil {
		return nil, err
	}

	return &remoteWriter{
		writer,
		location,
	}, nil
}

type remoteWriter struct {
	writer *writer
	location string
}

func (r remoteWriter) Write(p []byte) (n int, err error) {

}

func (r remoteWriter) Close() error {

}

func (r remoteWriter) Digest() digest.Digest {

}

func (r remoteWriter) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...content.Opt) error {
	return r.writer.commitBlob(r.location, expected.String())
}

func (r remoteWriter) Status() (content.Status, error) {

}

func (r remoteWriter) Truncate(size int64) error {

}
