package api

import (
	"github.com/simila-io/simila/api/gen/index/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"testing"
)

func TestJson2ProtoIndexRecord(t *testing.T) {
	irjs := `{
		"id": "dGVzdCB2YWx1ZQ==",
		"segment": "hello world",
		"vector": ["any string", "1234"]
	}`
	var r index.Record
	assert.Nil(t, protojson.Unmarshal([]byte(irjs), &r))
	cijs := `{
		"id": "dGVzdCB2YWx1ZQ==",
		"format": "pdf", 
		"records": [{"segment": "la la la", "vector": ["dd", "ff"]}] 
	}`
	var ci index.CreateIndexRequest
	assert.Nil(t, protojson.Unmarshal([]byte(cijs), &ci))
}
