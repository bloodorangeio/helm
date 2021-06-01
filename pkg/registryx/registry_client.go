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
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/oras-project/oras-go/pkg/auth"
	dockerauth "github.com/oras-project/oras-go/pkg/auth/docker"
	"helm.sh/helm/v3/pkg/helmpath"
)

type (
	// Client works with OCI-compliant registries
	Client struct {
		debug bool
		// path to repository config file e.g. ~/.docker/config.json
		credentialsFile string
		out             io.Writer
		authorizer      auth.Client
		resolver        remotes.Resolver
	}
)

// NewClient returns a new registry client with config
func NewClient(options ...ClientOption) (*Client, error) {
	client := &Client{
		out: ioutil.Discard,
	}
	for _, option := range options {
		option(client)
	}
	// set defaults if fields are missing
	if client.credentialsFile == "" {
		client.credentialsFile = helmpath.CachePath("registry", CredentialsFileBasename)
	}
	if client.authorizer == nil {
		authClient, err := dockerauth.NewClient(client.credentialsFile)
		if err != nil {
			return nil, err
		}
		client.authorizer = authClient
	}
	if client.resolver == nil {
		resolver, err := client.authorizer.Resolver(context.Background(), http.DefaultClient, false)
		if err != nil {
			return nil, err
		}
		client.resolver = resolver
	}
	return client, nil
}
