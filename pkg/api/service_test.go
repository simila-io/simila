package api

import (
	"github.com/simila-io/simila/api/gen/index/v1"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNodes2Create(t *testing.T) {
	tags := map[string]string{"key": "val"}
	assert.Equal(t, []persistence.Node{{Path: "/", Name: "aaa"}, {Path: "/aaa", Name: "bbb", Tags: tags}},
		nodes2Create([]string{"aaa", "bbb"}, nil, tags, index.NodeType_FOLDER))
	assert.Equal(t, []persistence.Node{{Path: "/aaa", Name: "bbb", Tags: tags}},
		nodes2Create([]string{"aaa", "bbb"}, []persistence.Node{{Path: "/", Name: "aaa"}}, tags, index.NodeType_FOLDER))
	assert.Equal(t, []persistence.Node{},
		nodes2Create([]string{"aaa", "bbb"}, []persistence.Node{{Path: "/", Name: "aaa"}, {Path: "/aaa", Name: "bbb"}}, nil, index.NodeType_FOLDER))
}
