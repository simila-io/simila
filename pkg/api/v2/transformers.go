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
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v2"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func toApiFormat(mFmt persistence.Format) *format.Format {
	return &format.Format{Name: mFmt.ID}
}

func toModelFormat(aFmt *format.Format) persistence.Format {
	if aFmt == nil {
		return persistence.Format{}
	}
	return persistence.Format{ID: aFmt.Name}
}

func toModelIndexRecordFromApiRecord(nID int64, aRec *index.Record, defRankMul float64) persistence.IndexRecord {
	if aRec == nil {
		return persistence.IndexRecord{}
	}
	rm := float64(aRec.RankMultiplier)
	if rm < 1.0 {
		rm = defRankMul
	}
	return persistence.IndexRecord{
		ID:       aRec.Id,
		NodeID:   nID,
		Segment:  aRec.Segment,
		RankMult: rm,
		Vector:   aRec.Vector,
	}
}

func toModelIndexRecordsFromApiRecords(nID int64, r []*index.Record, defRankMul float64) []persistence.IndexRecord {
	if len(r) == 0 {
		return nil
	}
	res := make([]persistence.IndexRecord, len(r))
	for i, ir := range r {
		res[i] = toModelIndexRecordFromApiRecord(nID, ir, defRankMul)
	}
	return res
}

func toApiNodes(nodes []persistence.Node) []*index.Node {
	res := make([]*index.Node, len(nodes))
	for i, n := range nodes {
		res[i] = toApiNode(n)
	}
	return res
}

func toApiNode(node persistence.Node) *index.Node {
	t := index.NodeType_FOLDER
	if node.Flags&persistence.NODE_FLAG_DOCUMENT != 0 {
		t = index.NodeType_DOCUMENT
	}
	return &index.Node{
		Path: node.Path,
		Name: node.Name,
		Tags: node.Tags,
		Type: t,
	}
}

func toApiRecord(mRec persistence.IndexRecord) *index.Record {
	return &index.Record{
		Id:             mRec.ID,
		Segment:        mRec.Segment,
		Vector:         mRec.Vector,
		Format:         mRec.Format,
		RankMultiplier: float32(mRec.RankMult),
	}
}

func toApiRecords(mRecs []persistence.IndexRecord) []*index.Record {
	res := make([]*index.Record, len(mRecs))
	for i, r := range mRecs {
		res[i] = toApiRecord(r)
	}
	return res
}

func toApiSearchRecord(sr persistence.SearchQueryResultItem) *index.SearchRecordsResultItem {
	res := &index.SearchRecordsResultItem{}
	res.Record = toApiRecord(sr.IndexRecord)
	res.Path = sr.Path
	res.MatchedKeywords = sr.MatchedKeywordsList
	res.Score = &sr.Score
	return res
}

func toApiSearchRecords(srs []persistence.SearchQueryResultItem) []*index.SearchRecordsResultItem {
	res := make([]*index.SearchRecordsResultItem, len(srs))
	for i, sr := range srs {
		res[i] = toApiSearchRecord(sr)
	}
	return res
}

func protoTime2Time(pt *timestamppb.Timestamp) time.Time {
	if pt == nil {
		return time.Time{}
	}
	return pt.AsTime()
}
