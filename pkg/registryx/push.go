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
	"math/rand"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/oras-project/oras-go/pkg/content"
	"github.com/oras-project/oras-go/pkg/oras"
)

// PushChart uploads a chart to a registry.
func (c *Client) PushChart(chartBytes []byte, ref string, opts ...PushOption) error {
	operation := &pushOperation{
		provBytes: nil,
	}
	for _, opt := range opts {
		opt(operation)
	}
	store := content.NewMemoryStore()
	descriptor := store.Add("", HelmChartContentLayerMediaType, chartBytes)
	// TODO: put Chart.yaml JSON-ified into config
	config := store.Add("", HelmChartConfigMediaType, []byte(fmt.Sprintf("{\"random\": \"%d\"}", rand.Int())))
	layers := []ocispec.Descriptor{descriptor}
	if operation.provBytes != nil {
		provDescriptor := store.Add("", HelmChartProvenanceLayerMediaType, operation.provBytes)
		layers = append(layers, provDescriptor)
	}
	_, err := oras.Push(ctx(c.out, c.debug), c.resolver, ref, store, layers,
		oras.WithConfig(config), oras.WithNameValidation(nil))
	return err
}
