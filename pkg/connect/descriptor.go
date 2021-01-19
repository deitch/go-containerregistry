package connect

import (
	"fmt"

	"github.com/containerd/containerd/remotes"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type Descriptor struct {
	name     string
	resolver remotes.Resolver
	desc     v1.Descriptor
}

func (d Descriptor) ImageIndex() (v1.ImageIndex, error) {
	mediaType := d.desc.MediaType
	switch  mediaType {
	case types.DockerManifestSchema1, types.DockerManifestSchema1Signed:
		// We don't care to support schema 1 images:
		// https://github.com/google/go-containerregistry/issues/377
		return nil, newErrSchema1(mediaType)
	case types.OCIManifestSchema1, types.DockerManifestSchema2:
		// We want an index but the registry has an image, nothing we can do.
		return nil, fmt.Errorf("unexpected media type for ImageIndex(): %s; call Image() instead", mediaType)
	case types.OCIImageIndex, types.DockerManifestList:
		// These are expected.
	default:
		// We could just return an error here, but some registries (e.g. static
		// registries) don't set the Content-Type headers correctly, so instead...
		logs.Warn.Printf("Unexpected media type for ImageIndex(): %s", mediaType)
	}
	return d.remoteIndex(), nil
}

func (d Descriptor) Image() (v1.Image, error) {
	mediaType := d.desc.MediaType
	switch mediaType {
	case types.DockerManifestSchema1, types.DockerManifestSchema1Signed:
		// We don't care to support schema 1 images:
		// https://github.com/google/go-containerregistry/issues/377
		return nil, newErrSchema1(mediaType)
	case types.OCIImageIndex, types.DockerManifestList:
		// We want an image but the registry has an index, resolve it to an image.
		return d.remoteIndex().imageByPlatform(d.platform)
	case types.OCIManifestSchema1, types.DockerManifestSchema2:
	default:
		logs.Warn.Printf("Unexpected media type for Image(): %s", mediaType)
	}

	// Wrap the v1.Layers returned by this v1.Image in a hint for downstream
	// remote.Write calls to facilitate cross-repo "mounting".
	imgCore, err := partial.CompressedToImage(d.remoteImage())
	if err != nil {
		return nil, err
	}
	return &mountableImage{
		Image:     imgCore,
		Reference: d.Ref,
	}
}

func (d Descriptor) Descriptor() v1.Descriptor {
	return d.desc
}
