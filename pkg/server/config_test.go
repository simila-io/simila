// Copyright 2023 The Simila Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package server

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildConfig_nofile(t *testing.T) {
	cfg, err := BuildConfig("")
	assert.Nil(t, err)
	assert.Equal(t, getDefaultConfig(), cfg)
}

func TestBuildConfig_file(t *testing.T) {
	dir, err := ioutil.TempDir("", "badTest")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	fn := filepath.Join(dir, "config.yaml")
	createFile(fn, `
grpctransport:
  network: "hoho"`)

	cfg, err := BuildConfig(fn)
	assert.Nil(t, err)
	assert.Equal(t, "hoho", cfg.GrpcTransport.Network)
}

func createFile(name, data string) {
	f, _ := os.Create(name)
	f.WriteString(data)
	f.Close()
}
