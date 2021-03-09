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
	"runtime"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
)

func TestCollectPlugins(t *testing.T) {
	env := cli.New()
	env.PluginsDirectory = pluginDir

	p, err := collectPlugins(env)
	if err != nil {
		t.Fatal(err)
	}

	if len(p) != 2 {
		t.Errorf("Expected 2 plugins, got %d: %v", len(p), p)
	}

	if _, err := p.ByScheme("test2"); err != nil {
		t.Error(err)
	}

	if _, err := p.ByScheme("test"); err != nil {
		t.Error(err)
	}

	if _, err := p.ByScheme("nosuchthing"); err == nil {
		t.Fatal("did not expect protocol handler for nosuchthing")
	}
}

func TestPluginPusher(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("TODO: refactor this test to work on windows")
	}

	env := cli.New()
	env.PluginsDirectory = pluginDir
	pp := NewPluginPusher("echo", env, "test", ".")
	_, err := pp()
	if err != nil {
		t.Fatal(err)
	}

	// TODO: test the Push interface here
}

func TestPluginSubCommands(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("TODO: refactor this test to work on windows")
	}

	env := cli.New()
	env.PluginsDirectory = pluginDir

	pp := NewPluginPusher("echo -n", env, "test", ".")
	_, err := pp()
	if err != nil {
		t.Fatal(err)
	}

	// TODO: test the Push interface here
}
