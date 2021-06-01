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

type (
	pushResult struct {
		Chart         *descriptorPushSummary
		Prov          *descriptorPushSummary
		Manifest      *manifestPushSummary
		Ref           string
		RefWithDigest string
	}

	descriptorPushSummary struct {
		Digest string
		Size   int64
	}

	manifestPushSummary struct {
		Digest string
		Size   int64
	}
)
