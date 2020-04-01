// Copyright 2020 Trey Dockendorf
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"testing"
)

func TestReloadConfigDefaults(t *testing.T) {
	sc := &SafeConfig{}
	err := sc.ReloadConfig("testdata/eseries_exporter.yaml")
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
		return
	}
	module, ok := sc.C.Modules["default"]
	if !ok {
		t.Errorf("Module default not loaded")
		return
	}
	if module.User != "monitor" {
		t.Errorf("Module User does not match monitor")
	}
}

func TestReloadConfigBadConfigs(t *testing.T) {
	sc := &SafeConfig{}
	tests := []struct {
		ConfigFile    string
		ExpectedError string
	}{
		{
			ConfigFile:    "/dne",
			ExpectedError: "Error reading config file /dne: open /dne: no such file or directory",
		},
		{
			ConfigFile:    "testdata/unknown-field.yaml",
			ExpectedError: "Error parsing config file testdata/unknown-field.yaml: yaml: unmarshal errors:\n  line 6: field invalid_extra_field not found in type config.Module",
		},
		{
			ConfigFile:    "testdata/missing-proxy-url.yaml",
			ExpectedError: "Module default must define 'proxy_url' value",
		},
		{
			ConfigFile:    "testdata/missing-user.yaml",
			ExpectedError: "Module default must define 'user' value",
		},
		{
			ConfigFile:    "testdata/missing-password.yaml",
			ExpectedError: "Module default must define 'password' value",
		},
	}
	for i, test := range tests {
		err := sc.ReloadConfig(test.ConfigFile)
		if err == nil {
			t.Errorf("In case %v:\nExpected:\n%v\nGot:\nnil", i, test.ExpectedError)
			continue
		}
		if err.Error() != test.ExpectedError {
			t.Errorf("In case %v:\nExpected:\n%v\nGot:\n%v", i, test.ExpectedError, err.Error())
		}
	}
}
