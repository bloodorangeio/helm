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

// Logout logs out of a registry
func (c *Client) Logout(host string, opts ...LogoutOption) (*logoutResult, error) {
	operation := &logoutOperation{}
	for _, opt := range opts {
		opt(operation)
	}
	err := c.authorizer.Logout(ctx(c.out, c.debug), host)
	if err != nil {
		return nil, err
	}
	result := &logoutResult{
		Host: host,
	}
	return result, nil
}
