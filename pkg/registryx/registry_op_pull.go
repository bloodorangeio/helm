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

// Pull downloads a chart from a registry
func (c *Client) Pull(ref string, options ...pullOption) (*pullResult, error) {
	operation := &pullOperation{
		withChart: true, // By default, always download the chart layer
	}
	for _, option := range options {
		option(operation)
	}
	if !operation.withChart && !operation.withProv {
		return nil, errors.New(
			"must specify at least one layer to pull (chart/prov)")
	}
	store := content.NewMemoryStore()
	allowedMediaTypes := []string{
		HelmChartConfigMediaType,
	}
	minNumLayers := 0
	if operation.withChart {
		minNumLayers += 1
		allowedMediaTypes = append(allowedMediaTypes, HelmChartContentLayerMediaType)
	}
	if operation.withProv {
		if !operation.ignoreMissingProv {
			minNumLayers += 1
		}
		allowedMediaTypes = append(allowedMediaTypes, HelmChartProvenanceLayerMediaType)
	}
	manifest, layerDescriptors, err := oras.Pull(ctx(c.out, c.debug), c.resolver, ref, store,
		oras.WithPullEmptyNameAllowed(),
		oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return nil, err
	}
	numLayers := len(layerDescriptors)
	if numLayers < minNumLayers {
		s := ""
		if minNumLayers > 1 {
			s = "s"
		}
		return nil, errors.New(
			fmt.Sprintf("manifest does not contain at least %d layer%s (total: %d)",
				minNumLayers, s, numLayers))
	}
	var chartLayer *ocispec.Descriptor
	var provLayer *ocispec.Descriptor
	for _, layer := range layerDescriptors {
		layer := layer
		switch layer.MediaType {
		case HelmChartContentLayerMediaType:
			chartLayer = &layer
		case HelmChartProvenanceLayerMediaType:
			provLayer = &layer
		}
	}
	if operation.withChart && chartLayer == nil {
		return nil, errors.New(
			fmt.Sprintf("manifest does not contain a layer with mediatype %s",
				HelmChartContentLayerMediaType))
	}
	var provMissing bool
	if operation.withProv && provLayer == nil {
		if operation.ignoreMissingProv {
			provMissing = true
		} else {
			return nil, errors.New(
				fmt.Sprintf("manifest does not contain a layer with mediatype %s",
					HelmChartContentLayerMediaType))
		}
	}
	var chartData []byte
	if operation.withChart {
		var ok bool
		_, chartData, ok = store.Get(*chartLayer)
		if !ok {
			return nil, errors.Errorf("Unable to retrieve blob with digest %s", chartLayer.Digest)
		}
	}
	var provData []byte
	if operation.withProv && !provMissing {
		var ok bool
		_, provData, ok = store.Get(*provLayer)
		if !ok {
			return nil, errors.Errorf("Unable to retrieve blob with digest %s", provLayer.Digest)
		}
	}
	result := &pullResult{
		Chart: &descriptorPullSummary{},
		Prov:  &descriptorPullSummary{},
		Manifest: &manifestPullSummary{
			Digest: manifest.Digest.String(),
			Size:   manifest.Size,
		},
	}
	if chartData != nil {
		result.Chart = &descriptorPullSummary{
			Data:   chartData,
			Digest: chartLayer.Digest.String(),
			Size:   chartLayer.Size,
		}
	}
	if provData != nil {
		result.Prov = &descriptorPullSummary{
			Data:   provData,
			Digest: provLayer.Digest.String(),
			Size:   provLayer.Size,
		}
	}
	return result, nil
}
