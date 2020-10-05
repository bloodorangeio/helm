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

package putter

import (
	"fmt"
	"strings"

	"helm.sh/helm/v3/internal/experimental/registry"
)

// OCIPusher is the default HTTP(/S) backend handler
type OCIPusher struct {
	opts options
}

//Push performs a Push from repo.Pusher and returns the body.
func (g *OCIPusher) Push(href string, options ...Option) error {
	for _, opt := range options {
		opt(&g.opts)
	}
	return g.push(href)
}

func (g *OCIPusher) push(href string) error {
	client := g.opts.registryClient

	ref := strings.TrimPrefix(href, "oci://")
	if version := g.opts.version; version != "" {
		ref = fmt.Sprintf("%s:%s", ref, version)
	}

	r, err := registry.ParseReference(ref)
	if err != nil {
		return err
	}

	return client.PushChart(r)
}

// NewOCIPusher constructs a valid http/https client as a Pusher
func NewOCIPusher(options ...Option) (Pusher, error) {
	var client OCIPusher

	for _, opt := range options {
		opt(&client.opts)
	}

	return &client, nil
}
