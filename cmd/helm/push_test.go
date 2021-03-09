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

package main

import (
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/repo/repotest"
)

func TestPushCmd(t *testing.T) {
	srv, err := repotest.NewTempServerWithCleanup(t, "testdata/testcharts/*.tgz*")
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	os.Setenv("HELM_EXPERIMENTAL_OCI", "1")
	ociSrv, err := repotest.NewOCIServer(t, srv.Root())
	if err != nil {
		t.Fatal(err)
	}
	ociSrv.Run(t)

	// TODO: this test
}

func TestPushVersionCompletion(t *testing.T) {
	// TODO: this test (?)
}

func TestPushFileCompletion(t *testing.T) {
	// TODO: this test
}
