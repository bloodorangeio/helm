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

<<<<<<< HEAD:pkg/registry/registry_op_push_options.go
type (
	// PushOption allows specifying various settings on push
	PushOption func(*pushOperation)
=======
import (
	"io"

	"helm.sh/helm/v3/internal/experimental/registry"
)

// ChartList performs a chart list operation.
type ChartList struct {
	cfg         *Configuration
	ColumnWidth uint
}
>>>>>>> bf486a25cdc12017c7dac74d1582a8a16acd37ea:pkg/action/chart_list.go

	pushOperation struct {
		provData []byte
	}
)

<<<<<<< HEAD:pkg/registry/registry_op_push_options.go
// PushOptProvData returns a function that sets the prov bytes setting on push
func PushOptProvData(provData []byte) PushOption {
	return func(operation *pushOperation) {
		operation.provData = provData
	}
=======
// Run executes the chart list operation
func (a *ChartList) Run(out io.Writer) error {
	client := a.cfg.RegistryClient
	opt := registry.ClientOptColumnWidth(a.ColumnWidth)
	opt(client)
	return client.PrintChartTable()
>>>>>>> bf486a25cdc12017c7dac74d1582a8a16acd37ea:pkg/action/chart_list.go
}
