package connect

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/containerd/containerd/content"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"golang.org/x/sync/errgroup"
)

// Write given a Descriptor, write it to the ingestor
func Write(r Descriptor, ingester content.Ingester) error {
	desc := r.Descriptor()
	ii := r.AsIndex()
	if ii != nil {
		return writeIndex(ii, desc, ingester)
	}
	img := r.AsImage()
	if img != nil {
		return writeImage(img, desc, ingester)
	}
	return errors.New("descriptor was neither image nor index")
}

func writeIndex(ii v1.ImageIndex, desc v1.Descriptor, ingester content.Ingester) error {
	index, err := ii.IndexManifest()
	if err != nil {
		return err
	}
	for _, v1desc := range index.Manifests {
		switch v1desc.MediaType {
		case types.OCIImageIndex, types.DockerManifestList:
			ii, err := ii.ImageIndex(v1desc.Digest)
			if err != nil {
				return err
			}
			if err := writeIndex(ii, v1desc, ingester); err != nil {
				return err
			}
		case types.OCIManifestSchema1, types.DockerManifestSchema2:
			img, err := ii.Image(v1desc.Digest)
			if err != nil {
				return err
			}
			if err := writeImage(img, v1desc, ingester); err != nil {
				return err
			}
		default:

		}
	}
	return nil
}

func writeBlob(ctx context.Context, rc io.ReadCloser, v1desc v1.Descriptor, ingester content.Ingester) error {
	desc, err := v1desc.Ocispec()
	if err != nil {
		return err
	}
	writer, err := ingester.Writer(ctx, content.WithDescriptor(desc))
	n, err := io.Copy(writer, rc)
	if err != nil {
		return err
	}
	if n != desc.Size {
		return fmt.Errorf("mismatched sizes, actual %d, expected %d", n, desc.Size)
	}
	return writer.Commit(ctx, desc.Size, desc.Digest)
}

func writeImage(img v1.Image, desc v1.Descriptor, ingester content.Ingester) error {
	ctx := context.Background()
	manifest, err := img.Manifest()
	if err != nil {
		return err
	}
	layers, err := img.Layers()
	if err != nil {
		return err
	}

	// Write the layers concurrently.
	var g errgroup.Group
	for i, layer := range layers {
		layer := layer
		g.Go(func() error {
			rc, err := layer.Compressed()
			if err != nil {
				return err
			}
			return writeBlob(ctx, rc, manifest.Layers[i], ingester)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	// Write the config.
	cfgBlob, err := img.RawConfigFile()
	if err != nil {
		return err
	}
	if err := writeBlob(ctx, ioutil.NopCloser(bytes.NewReader(cfgBlob)), manifest.Config, ingester); err != nil {
		return err
	}

	// Write the img manifest.
	rawManifest, err := img.RawManifest()
	if err != nil {
		return err
	}
	if err := writeBlob(ctx, ioutil.NopCloser(bytes.NewReader(rawManifest)), desc, ingester); err != nil {
		return err
	}
	return nil
}
