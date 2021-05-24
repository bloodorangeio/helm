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

package registry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/containerd/containerd/errdefs"
	auth "github.com/oras-project/oras-go/pkg/auth/docker"
	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/registry"
	_ "github.com/docker/distribution/registry/auth/htpasswd"
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

var (
	testWorkspaceRootDir     = "helm-registry-test"
	testHtpasswdFileBasename = "authtest.htpasswd"
	testUsername             = "myuser"
	testPassword             = "mypass"
)

type RegistryClientTestSuite struct {
	suite.Suite
	Out                     io.Writer
	DockerRegistryHost      string
	CompromisedRegistryHost string
	WorkspaceRootDir        string
	RegistryClient          *Client
}

func (suite *RegistryClientTestSuite) SetupSuite() {
	suite.WorkspaceRootDir = filepath.Join("testdata", testWorkspaceRootDir)
	os.RemoveAll(suite.WorkspaceRootDir)
	os.Mkdir(suite.WorkspaceRootDir, 0700)

	var out bytes.Buffer
	suite.Out = &out
	credentialsFile := filepath.Join(suite.WorkspaceRootDir, CredentialsFileBasename)

	client, err := auth.NewClient(credentialsFile)
	suite.Nil(err, "no error creating auth client")

	resolver, err := client.Resolver(context.Background(), http.DefaultClient, false)
	suite.Nil(err, "no error creating resolver")

	// init test client
	suite.RegistryClient, err = NewClient(
		ClientOptDebug(true),
		ClientOptWriter(suite.Out),
		ClientOptAuthorizer(&Authorizer{
			Client: client,
		}),
		ClientOptResolver(&Resolver{
			Resolver: resolver,
		}),
	)
	suite.Nil(err, "no error creating registry client")

	// create htpasswd file (w BCrypt, which is required)
	pwBytes, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	suite.Nil(err, "no error generating bcrypt password for test htpasswd file")
	htpasswdPath := filepath.Join(suite.WorkspaceRootDir, testHtpasswdFileBasename)
	err = ioutil.WriteFile(htpasswdPath, []byte(fmt.Sprintf("%s:%s\n", testUsername, string(pwBytes))), 0644)
	suite.Nil(err, "no error creating test htpasswd file")

	// Registry config
	config := &configuration.Configuration{}
	port, err := freeport.GetFreePort()
	suite.Nil(err, "no error finding free port for test registry")
	suite.DockerRegistryHost = fmt.Sprintf("localhost:%d", port)
	config.HTTP.Addr = fmt.Sprintf(":%d", port)
	config.HTTP.DrainTimeout = time.Duration(10) * time.Second
	config.Storage = map[string]configuration.Parameters{"inmemory": map[string]interface{}{}}
	config.Auth = configuration.Auth{
		"htpasswd": configuration.Parameters{
			"realm": "localhost",
			"path":  htpasswdPath,
		},
	}
	dockerRegistry, err := registry.NewRegistry(context.Background(), config)
	suite.Nil(err, "no error creating test registry")

	suite.CompromisedRegistryHost = initCompromisedRegistryTestServer()

	// Start Docker registry
	go dockerRegistry.ListenAndServe()
}

func (suite *RegistryClientTestSuite) TearDownSuite() {
	os.RemoveAll(suite.WorkspaceRootDir)
}

func (suite *RegistryClientTestSuite) Test_0_Login() {
	err := suite.RegistryClient.Login(suite.DockerRegistryHost, "badverybad", "ohsobad", false)
	suite.NotNil(err, "error logging into registry with bad credentials")

	err = suite.RegistryClient.Login(suite.DockerRegistryHost, "badverybad", "ohsobad", true)
	suite.NotNil(err, "error logging into registry with bad credentials, insecure mode")

	err = suite.RegistryClient.Login(suite.DockerRegistryHost, testUsername, testPassword, false)
	suite.Nil(err, "no error logging into registry with good credentials")

	err = suite.RegistryClient.Login(suite.DockerRegistryHost, testUsername, testPassword, true)
	suite.Nil(err, "no error logging into registry with good credentials, insecure mode")
}

func (suite *RegistryClientTestSuite) Test_1_PushChart() {
	contentBytes, err := ioutil.ReadFile("testdata/mychart")
	suite.NotNil(err)

	ref, err := ParseReference(fmt.Sprintf("%s/testrepo/mychart:0.1.0", suite.DockerRegistryHost))
	suite.Nil(err)

	err = suite.RegistryClient.PushChart(contentBytes, ref)
	suite.Nil(err)
}

func (suite *RegistryClientTestSuite) Test_2_PullChart() {

	// non-existent ref
	ref, err := ParseReference(fmt.Sprintf("%s/testrepo/whodis:9.9.9", suite.DockerRegistryHost))
	suite.Nil(err)
	_, err = suite.RegistryClient.PullChart(ref)
	suite.NotNil(err)

	// existing ref
	ref, err = ParseReference(fmt.Sprintf("%s/testrepo/mychart:0.1.0", suite.DockerRegistryHost))
	suite.Nil(err)
	_, err = suite.RegistryClient.PullChart(ref)
	suite.Nil(err)
}

func (suite *RegistryClientTestSuite) Test_3_Logout() {
	err := suite.RegistryClient.Logout("this-host-aint-real:5000")
	suite.NotNil(err, "error logging out of registry that has no entry")

	err = suite.RegistryClient.Logout(suite.DockerRegistryHost)
	suite.Nil(err, "no error logging out of registry")
}

func (suite *RegistryClientTestSuite) Test_4_ManInTheMiddle() {
	ref, err := ParseReference(fmt.Sprintf("%s/testrepo/supposedlysafechart:9.9.9", suite.CompromisedRegistryHost))
	suite.Nil(err)

	// returns content that does not match the expected digest
	_, err = suite.RegistryClient.PullChart(ref)
	suite.NotNil(err)
	fmt.Println(err)
	suite.True(errdefs.IsFailedPrecondition(err))
}

func TestRegistryClientTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryClientTestSuite))
}

func initCompromisedRegistryTestServer() string {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "manifests") {
			w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
			w.WriteHeader(200)

			// layers[0] is the blob []byte("a")
			w.Write([]byte(
				`{ "schemaVersion": 2, "config": {
    "mediaType": "application/vnd.cncf.helm.config.v1+json",
    "digest": "sha256:a705ee2789ab50a5ba20930f246dbd5cc01ff9712825bb98f57ee8414377f133",
    "size": 181
  },
  "layers": [
    {
      "mediaType": "application/vnd.cncf.helm.chart.content.v1.tar+gzip",
      "digest": "sha256:ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
      "size": 1
    }
  ]
}`))
		} else if r.URL.Path == "/v2/testrepo/supposedlysafechart/blobs/sha256:a705ee2789ab50a5ba20930f246dbd5cc01ff9712825bb98f57ee8414377f133" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte("{\"name\":\"mychart\",\"version\":\"0.1.0\",\"description\":\"A Helm chart for Kubernetes\\n" +
				"an 'application' or a 'library' chart.\",\"apiVersion\":\"v2\",\"appVersion\":\"1.16.0\",\"type\":" +
				"\"application\"}"))
		} else if r.URL.Path == "/v2/testrepo/supposedlysafechart/blobs/sha256:ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb" {
			w.Header().Set("Content-Type", "application/vnd.cncf.helm.chart.content.v1.tar+gzip")
			w.WriteHeader(200)
			w.Write([]byte("b"))
		} else {
			w.WriteHeader(500)
		}
	}))

	u, _ := url.Parse(s.URL)
	return fmt.Sprintf("localhost:%s", u.Port())
}
