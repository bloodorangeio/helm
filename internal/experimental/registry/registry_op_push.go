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

package registry // import "helm.sh/helm/v3/internal/experimental/registry"

import (
	"encoding/json"
	
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/oras-project/oras-go/pkg/content"
	"github.com/oras-project/oras-go/pkg/oras"
)

// Push uploads a chart to a registry.
func (c *Client) Push(data []byte, ref string, options ...PushOption) (*pushResult, error) {
	operation := &pushOperation{}
	for _, option := range options {
		option(operation)
	}
	store := content.NewMemoryStore()
	descriptor := store.Add("", ChartLayerMediaType, data)
	meta, err := extractChartMeta(data)
	if err != nil {
		return nil, err
	}
	configData, err := json.Marshal(meta)
	config := store.Add("", ConfigMediaType, configData)
	layers := []ocispec.Descriptor{descriptor}
	var provDescriptor ocispec.Descriptor
	if operation.provData != nil {
		provDescriptor = store.Add("", ProvLayerMediaType, operation.provData)
		layers = append(layers, provDescriptor)
	}
	manifest, err := oras.Push(ctx(c.out, c.debug), c.resolver, ref, store, layers,
		oras.WithConfig(config), oras.WithNameValidation(nil))
	if err != nil {
		return nil, err
	}
	chartSummary := &descriptorPushSummaryWithMeta{
		Meta: meta,
	}
	chartSummary.Digest = descriptor.Digest.String()
	chartSummary.Size = descriptor.Size
	result := &pushResult{
		Manifest: &descriptorPushSummary{
			Digest: manifest.Digest.String(),
			Size:   manifest.Size,
		},
		Config: &descriptorPushSummary{
			Digest: config.Digest.String(),
			Size:   config.Size,
		},
		Chart: chartSummary,
		Prov:  &descriptorPushSummary{}, // prevent nil references
		Ref:   ref,
	}
	if operation.provData != nil {
		result.Prov = &descriptorPushSummary{
			Digest: provDescriptor.Digest.String(),
			Size:   provDescriptor.Size,
		}
	}
	return result, err
}
