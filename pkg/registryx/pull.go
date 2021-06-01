/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package registryx // import "helm.sh/helm/v3/pkg/registry"

import (
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/oras-project/oras-go/pkg/content"
	"github.com/oras-project/oras-go/pkg/oras"
	"github.com/pkg/errors"
)

// PullChart downloads a chart from a registry
func (c *Client) PullChart(ref string, opts ...PullOption) ([]byte, error) {
	return c.pullByMediaType(ref, HelmChartContentLayerMediaType, opts...)
}

// PullProv downloads a provenance file from a registry
func (c *Client) PullProv(ref string, opts ...PullOption) ([]byte, error) {
	return c.pullByMediaType(ref, HelmChartProvenanceLayerMediaType, opts...)
}

func (c *Client) pullByMediaType(ref string, mediaType string, opts ...PullOption) ([]byte, error) {
	operation := &pullOperation{}
	for _, opt := range opts {
		opt(operation)
	}
	store := content.NewMemoryStore()
	_, layerDescriptors, err := oras.Pull(ctx(c.out, c.debug), c.resolver, ref, store,
		oras.WithPullEmptyNameAllowed(),
		oras.WithAllowedMediaTypes([]string{
			HelmChartConfigMediaType,
			mediaType,
		}))
	if err != nil {
		return nil, err
	}
	numLayers := len(layerDescriptors)
	if numLayers < 1 {
		return nil, errors.New(
			fmt.Sprintf("manifest does not contain at least 1 layers (total: %d)", numLayers))
	}
	var desiredLayer *ocispec.Descriptor
	for _, layer := range layerDescriptors {
		layer := layer
		switch layer.MediaType {
		case mediaType:
			desiredLayer = &layer
		}
	}
	if desiredLayer == nil {
		return nil, errors.New(
			fmt.Sprintf("manifest does not contain a layer with mediatype %s",
				mediaType))
	}
	_, b, ok := store.Get(*desiredLayer)
	if !ok {
		return nil, errors.Errorf("Unable to retrieve blob with digest %s", desiredLayer.Digest)
	}
	return b, nil
}
