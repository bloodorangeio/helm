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

package action

import (
	"strings"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/pusher"
	"helm.sh/helm/v3/pkg/uploader"
)

// Push is the action for uploading a chart.
//
// It provides the implementation of 'helm push'.
type Push struct {
	ChartPathOptions

	Settings *cli.EnvSettings // TODO: refactor this out of pkg/action

	Devel       bool
	Untar       bool
	VerifyLater bool
	UntarDir    string
	DestDir     string
	cfg         *Configuration
}

type PushOpt func(*Push)

func WithPushConfig(cfg *Configuration) PushOpt {
	return func(p *Push) {
		p.cfg = cfg
	}
}

// NewPush creates a new Push object.
func NewPush() *Push {
	return NewPushWithOpts()
}

// NewPushWithOpts creates a new push, with configuration options.
func NewPushWithOpts(opts ...PushOpt) *Push {
	p := &Push{}
	for _, fn := range opts {
		fn(p)
	}

	return p
}

// Run executes 'helm push' against the given chart archive.
func (p *Push) Run(chartRef string, remote string) (string, error) {
	var out strings.Builder

	c := uploader.ChartUploader{
		Out:     &out,
		Pushers: pusher.All(p.Settings),
		Options: []pusher.Option{
			pusher.WithBasicAuth(p.Username, p.Password),
			pusher.WithTLSClientConfig(p.CertFile, p.KeyFile, p.CaFile),
			pusher.WithInsecureSkipVerifyTLS(p.InsecureSkipTLSverify),
		},
	}

	if strings.HasPrefix(remote, "oci://") {
		c.Options = append(c.Options, pusher.WithRegistryClient(p.cfg.RegistryClient))
		if p.Version != "" {
			// TODO: rename to WithVersion ?
			c.Options = append(c.Options, pusher.WithTagName(p.Version))
		}
	}

	err := c.UploadTo(chartRef, remote)
	return out.String(), err
}
