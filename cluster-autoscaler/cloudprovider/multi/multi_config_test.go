/*
Copyright 2023 The Kubernetes Authors.

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

package multi

import "testing"

func TestNewConfig(t *testing.T) {
	cfg, err := newConfig("./examples/config-with-cloudconfig.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Providers) == 0 {
		t.Fatal("expected providers to be set")
	}

	if len(cfg.Providers[0].Name) == 0 {
		t.Fatal("expected name of the first provider to be set")
	}
	
	if len(cfg.Providers[0].CloudConfig) == 0 {
		t.Fatal("expected cloud-config of the first provider to be set")
	}

	if len(cfg.Providers[1].Name) == 0 {
		t.Fatal("expected name of the second provider to be set")
	}
}
