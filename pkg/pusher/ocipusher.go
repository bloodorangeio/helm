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

package pusher

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"helm.sh/helm/v3/internal/experimental/registry"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// OCIPusher is the default OCI backend handler
type OCIPusher struct {
	opts options
}

// Push performs a Push from repo.Pusher.
func (g *OCIPusher) Push(chartRef, href string, options ...Option) error {
	for _, opt := range options {
		opt(&g.opts)
	}
	return g.push(chartRef, href)
}

func (g *OCIPusher) push(chartRef, href string) error {
	meta, err := loader.Load(chartRef)
	if err != nil {
		return err
	}

	client := g.opts.registryClient

	ref := strings.TrimPrefix(href, "oci://")
	if version := g.opts.version; version != "" {
		ref = fmt.Sprintf("%s:%s", ref, version)
	} else {
		ref = fmt.Sprintf("%s:%s", ref, meta.Metadata.Version)
	}

	r, err := registry.ParseReference(ref)
	if err != nil {
		return err
	}
	r.Repo = path.Join(r.Repo, meta.Metadata.Name)

	bytes, err := ioutil.ReadFile(chartRef)
	if err != nil {
		return err
	}
	err = client.PushChart(bytes, r)
	if err != nil {
		return err
	}

	return nil
}

// NewOCIPusher constructs a valid OCI client as a Pusher
func NewOCIPusher(ops ...Option) (Pusher, error) {
	registryClient, err := registry.NewClient()
	if err != nil {
		return nil, err
	}

	client := OCIPusher{
		opts: options{
			registryClient: registryClient,
		},
	}

	for _, opt := range ops {
		opt(&client.opts)
	}

	return &client, nil
}
