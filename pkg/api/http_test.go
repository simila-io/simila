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
package api

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJson2ProtoIndexRecord(t *testing.T) {
	irjs := `{
		"id": "dGVzdCB2YWx1ZQ==",
		"segment": "hello world",
		"vector": ["any string", "1234"]
	}`
	var r record
	assert.Nil(t, json.Unmarshal([]byte(irjs), &r))
	cijs := `{
		"id": "dGVzdCB2YWx1ZQ==",
		"format": "pdf", 
		"records": [{"segment": "la la la", "vector": ["dd", "ff"]}] 
	}`
	var ci createIndexRequest
	assert.Nil(t, json.Unmarshal([]byte(cijs), &ci))
}
