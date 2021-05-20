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

package registry // import "helm.sh/helm/v3/pkg/registry"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	auth "github.com/oras-project/oras-go/pkg/auth/docker"
	"github.com/oras-project/oras-go/pkg/content"
	"github.com/oras-project/oras-go/pkg/oras"
	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/helmpath"
)

const (
	// CredentialsFileBasename is the filename for auth credentials file
	CredentialsFileBasename = "config.json"
)

type (
	// Client works with OCI-compliant registries
	Client struct {
		debug bool
		// path to repository config file e.g. ~/.docker/config.json
		credentialsFile string
		out             io.Writer
		authorizer      *Authorizer
		resolver        *Resolver
	}
)

// NewClient returns a new registry client with config
func NewClient(opts ...ClientOption) (*Client, error) {
	client := &Client{
		out: ioutil.Discard,
	}
	for _, opt := range opts {
		opt(client)
	}
	// set defaults if fields are missing
	if client.credentialsFile == "" {
		client.credentialsFile = helmpath.CachePath("registry", CredentialsFileBasename)
	}
	if client.authorizer == nil {
		authClient, err := auth.NewClient(client.credentialsFile)
		if err != nil {
			return nil, err
		}
		client.authorizer = &Authorizer{
			Client: authClient,
		}
	}
	if client.resolver == nil {
		resolver, err := client.authorizer.Resolver(context.Background(), http.DefaultClient, false)
		if err != nil {
			return nil, err
		}
		client.resolver = &Resolver{
			Resolver: resolver,
		}
	}
	return client, nil
}

// Login logs into a registry
func (c *Client) Login(hostname string, username string, password string, insecure bool) error {
	err := c.authorizer.Login(ctx(c.out, c.debug), hostname, username, password, insecure)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "Login succeeded\n")
	return nil
}

// Logout logs out of a registry
func (c *Client) Logout(hostname string) error {
	err := c.authorizer.Logout(ctx(c.out, c.debug), hostname)
	if err != nil {
		return err
	}
	fmt.Fprintln(c.out, "Logout succeeded")
	return nil
}

// PushChart uploads a chart to a registry.
func (c *Client) PushChart(chartBytes []byte, provBytes []byte, ref *Reference) error {
	fmt.Fprintf(c.out, "The push refers to repository [%s]\n", ref.Repo)

	store := content.NewMemoryStore()
	descriptor := store.Add("", HelmChartContentLayerMediaType, chartBytes)

	// TODO: put Chart.yaml JSON-ified into config
	config := store.Add("", HelmChartConfigMediaType, []byte("{}"))

	layers := []ocispec.Descriptor{descriptor}
	if provBytes != nil {
		provDescriptor := store.Add("", HelmChartProvenanceLayerMediaType, provBytes)
		layers = append(layers, provDescriptor)
	}

	_, err := oras.Push(ctx(c.out, c.debug), c.resolver, ref.FullName(), store, layers,
		oras.WithConfig(config), oras.WithNameValidation(nil))
	if err != nil {
		return err
	}
	s := ""
	numLayers := len(layers)
	if 1 < numLayers {
		s = "s"
	}

	// TODO: use actual size of content
	fmt.Fprintf(c.out,
		"%s: pushed to remote (%d layer%s, %s total)\n", ref.Tag, numLayers, s, byteCountBinary(1024))
	return nil
}

// PullChart downloads a chart from a registry
func (c *Client) PullChart(ref *Reference) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	if ref.Tag == "" {
		return buf, errors.New("tag explicitly required")
	}

	fmt.Fprintf(c.out, "%s: Pulling from %s\n", ref.Tag, ref.Repo)

	store := content.NewMemoryStore()
	fullname := ref.FullName()
	_ = fullname
	_, layerDescriptors, err := oras.Pull(ctx(c.out, c.debug), c.resolver, ref.FullName(), store,
		oras.WithPullEmptyNameAllowed(),
		oras.WithAllowedMediaTypes(KnownMediaTypes()))
	if err != nil {
		return buf, err
	}

	numLayers := len(layerDescriptors)
	if numLayers < 1 {
		return buf, errors.New(
			fmt.Sprintf("manifest does not contain at least 1 layer (total: %d)", numLayers))
	}

	var contentLayer *ocispec.Descriptor
	var provLayer *ocispec.Descriptor
	for _, layer := range layerDescriptors {
		layer := layer
		switch layer.MediaType {
		case HelmChartContentLayerMediaType:
			contentLayer = &layer
		case HelmChartProvenanceLayerMediaType:
			provLayer = &layer
		}
	}

	if contentLayer == nil {
		return buf, errors.New(
			fmt.Sprintf("manifest does not contain a layer with mediatype %s",
				HelmChartContentLayerMediaType))
	}

	_, b, ok := store.Get(*contentLayer)
	if !ok {
		return buf, errors.Errorf("Unable to retrieve blob with digest %s", contentLayer.Digest)
	}
	buf = bytes.NewBuffer(b)

	if provLayer != nil {
		_, bb, ok := store.Get(*provLayer)
		if !ok {
			return buf, errors.Errorf("Unable to retrieve blob with digest %s", contentLayer.Digest)
		}

		_, err = buf.Write([]byte("<--MYSTERIOUSDIVIDER-->"))
		if err != nil {
			return buf, err
		}

		_, err = buf.Write(bb)
		if err != nil {
			return buf, err
		}
	}

	return buf, nil
}
