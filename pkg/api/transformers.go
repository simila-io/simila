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
	"github.com/acquirecloud/golibs/cast"
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v1"
	similapi "github.com/simila-io/simila/api/genpublic/v1"
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
		Format:   aRec.Format,
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
	if node.Flags&persistence.NodeFlagDocument != 0 {
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

func format2Rest(f *format.Format) similapi.Format {
	return similapi.Format{Name: f.Name}
}

func formats2Rest(fs *format.Formats) similapi.Formats {
	res := similapi.Formats{}
	if fs == nil || len(fs.Formats) == 0 {
		return res
	}
	res.Formats = make([]similapi.Format, len(fs.Formats))
	for i, f := range fs.Formats {
		res.Formats[i] = format2Rest(f)
	}
	return res
}

func searchRecordsRequest2Proto(sr similapi.SearchRecordsRequest) *index.SearchRecordsRequest {
	return &index.SearchRecordsRequest{
		Text:   sr.Text,
		Format: sr.Format,
		Tags:   sr.Tags,
		Strict: cast.Ptr(sr.Strict),
		Path:   sr.Path,
		Offset: cast.Ptr(int64(sr.Offset)),
		Limit:  cast.Ptr(int64(sr.Limit)),
	}
}

func searchRecordsResult2Rest(srr *index.SearchRecordsResult) similapi.SearchRecordsResult {
	res := similapi.SearchRecordsResult{}
	if srr == nil {
		return res
	}
	res.Items = make([]similapi.SearchRecordsResultItem, len(srr.Items))
	for i, sr := range srr.Items {
		res.Items[i] = searchRecordsResultItems2Rest(sr)
	}
	res.Total = int(srr.Total)
	return res
}

func searchRecordsResultItems2Rest(srr *index.SearchRecordsResultItem) similapi.SearchRecordsResultItem {
	return similapi.SearchRecordsResultItem{
		Record:          record2Rest(srr.Record),
		Path:            srr.Path,
		Score:           cast.Value(srr.Score, -1.0),
		MatchedKeywords: srr.MatchedKeywords,
	}
}

func records2Rest(rs []*index.Record) []similapi.Record {
	if len(rs) == 0 {
		return nil
	}
	res := make([]similapi.Record, len(rs))
	for i, r := range rs {
		res[i] = record2Rest(r)
	}
	return res
}

func record2Rest(r *index.Record) similapi.Record {
	return similapi.Record{
		Id:             r.Id,
		Segment:        r.Segment,
		Vector:         r.Vector,
		RankMultiplier: r.RankMultiplier,
		Format:         r.Format,
	}
}

func node2Rest(n *index.Node) similapi.Node {
	if n == nil {
		return similapi.Node{}
	}
	tp := similapi.Folder
	if n.Type == index.NodeType_DOCUMENT {
		tp = similapi.Document
	}
	return similapi.Node{
		Path: n.Path,
		Name: n.Name,
		Tags: n.Tags,
		Type: tp,
	}
}

func nodes2Rest(ns []*index.Node) []similapi.Node {
	res := make([]similapi.Node, len(ns))
	for i, n := range ns {
		res[i] = node2Rest(n)
	}
	return res
}

func rest2Records(rs []similapi.Record) []*index.Record {
	res := make([]*index.Record, len(rs))
	for i, r := range rs {
		res[i] = rest2Record(r)
	}
	return res
}

func rest2Record(r similapi.Record) *index.Record {
	return &index.Record{
		Id:             r.Id,
		Format:         r.Format,
		Segment:        r.Segment,
		Vector:         r.Vector,
		RankMultiplier: r.RankMultiplier,
	}
}

func rest2CreateRecordsRequest(path string, crr similapi.CreateRecordsRequest) *index.CreateRecordsRequest {
	res := &index.CreateRecordsRequest{}
	res.Records = rest2Records(crr.Records)
	res.Path = path
	res.Tags = crr.Tags
	res.Document = crr.Document
	res.RankMultiplier = crr.RankMultiplier
	if crr.Parser != "" {
		res.Parser = cast.Ptr(crr.Parser)
	}
	res.NodeType = cast.Ptr(index.NodeType_DOCUMENT)
	if crr.NodeType == similapi.Folder {
		res.NodeType = cast.Ptr(index.NodeType_FOLDER)
	}
	return res
}

func protoTime2Time(pt *timestamppb.Timestamp) time.Time {
	if pt == nil {
		return time.Time{}
	}
	return pt.AsTime()
}
