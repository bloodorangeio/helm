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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	orascontext "github.com/oras-project/oras-go/pkg/context"
	"github.com/sirupsen/logrus"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// BuildRefFromChartData generates an OCI reference recommended for
// Helm charts stored in an OCI registry, in the form
// of <host>/<parent>/<chart_name>:<chart_version>
func BuildRefFromChartData(prefix string, chartData []byte) (string, error) {
	meta, err := extractChartMeta(chartData)
	if err != nil {
		return "", err
	}
	ref := fmt.Sprintf("%s:%s", path.Join(prefix, meta.Name), meta.Version)
	return ref, nil
}

// extractChartMeta is used to extract a chart metadata from a byte array
// that is loaded from a chart .tgz. Currently it is necessary to create a
// temporary file with the .tgz contents in order to use chartutil.LoadChartfile
func extractChartMeta(chartData []byte) (*chart.Metadata, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "helm-oci-extract-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(chartData)
	ch, err := loader.Load(tmpFile.Name())
	if err != nil {
		return nil, err
	}
	return ch.Metadata, nil
}

// ctx retrieves a fresh context.
// disable verbose logging coming from ORAS (unless debug is enabled)
func ctx(out io.Writer, debug bool) context.Context {
	if !debug {
		return orascontext.Background()
	}
	ctx := orascontext.WithLoggerFromWriter(context.Background(), out)
	orascontext.GetLogger(ctx).Logger.SetLevel(logrus.DebugLevel)
	return ctx
}
