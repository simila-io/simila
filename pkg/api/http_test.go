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
	var r Record
	assert.Nil(t, json.Unmarshal([]byte(irjs), &r))
	cijs := `{
		"id": "dGVzdCB2YWx1ZQ==",
		"format": "pdf", 
		"records": [{"segment": "la la la", "vector": ["dd", "ff"]}] 
	}`
	var ci CreateIndexRequest
	assert.Nil(t, json.Unmarshal([]byte(cijs), &ci))
}
