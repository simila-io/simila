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

func toApiIndex(mIdx persistence.Index) *index.Index {
	return &index.Index{
		Id:        mIdx.ID,
		Format:    mIdx.Format,
		Tags:      mIdx.Tags,
		CreatedAt: timestamppb.New(mIdx.CreatedAt),
	}
}

func toModelIndex(aIdx *index.Index) persistence.Index {
	if aIdx == nil {
		return persistence.Index{}
	}
	return persistence.Index{ID: aIdx.Id, Format: aIdx.Format, Tags: aIdx.Tags}
}

func toModelIndexFromApiCreateIdxReq(request *index.CreateIndexRequest) persistence.Index {
	if request == nil {
		return persistence.Index{}
	}
	return persistence.Index{ID: request.Id, Tags: request.Tags, Format: request.Format}
}

func toApiSearchResultItem(mItem persistence.SearchQueryResultItem, includeScore bool) *index.SearchRecordsResultItem {
	srr := &index.SearchRecordsResultItem{
		IndexId:         mItem.IndexID,
		IndexRecord:     toApiRecord(mItem.IndexRecord),
		MatchedKeywords: mItem.MatchedKeywordsList,
	}
	if includeScore {
		srr.Score = &mItem.Score
	}
	return srr
}

func toModelIndexRecordFromApiRecord(indexID string, aRec *index.Record) persistence.IndexRecord {
	if aRec == nil {
		return persistence.IndexRecord{}
	}
	return persistence.IndexRecord{
		ID:      aRec.Id,
		IndexID: indexID,
		Segment: aRec.Segment,
		Vector:  aRec.Vector,
	}
}

func toModelIndexRecordsFromApiRecords(idxId string, r []*index.Record) []persistence.IndexRecord {
	if len(r) == 0 {
		return nil
	}
	res := make([]persistence.IndexRecord, len(r))
	for i, ir := range r {
		res[i] = toModelIndexRecordFromApiRecord(idxId, ir)
	}
	return res
}

func toApiRecord(mRec persistence.IndexRecord) *index.Record {
	return &index.Record{
		Id:      mRec.ID,
		Segment: mRec.Segment,
		Vector:  mRec.Vector,
	}
}

func createIndexRequest2Proto(ci similapi.CreateIndexRequest) *index.CreateIndexRequest {
	return &index.CreateIndexRequest{
		Id:       ci.Id,
		Format:   ci.Format,
		Tags:     ci.Tags,
		Document: ci.Document,
		Records:  records2Proto(ci.Records),
	}
}

func records2Proto(rs []similapi.Record) []*index.Record {
	if len(rs) == 0 {
		return nil
	}
	res := make([]*index.Record, len(rs))
	for i, r := range rs {
		res[i] = record2Proto(r)
	}
	return res
}

func record2Proto(r similapi.Record) *index.Record {
	return &index.Record{
		Id:      r.Id,
		Segment: r.Segment,
		Vector:  r.Vector,
	}
}

func patchIndexRecordsRequest2Proto(pr similapi.PatchRecordsRequest) *index.PatchRecordsRequest {
	return &index.PatchRecordsRequest{
		Id:            pr.Id,
		UpsertRecords: records2Proto(pr.UpsertRecords),
		DeleteRecords: records2Proto(pr.DeleteRecords),
	}
}

func patchIndexRecordsResult2Rest(prr *index.PatchRecordsResult) similapi.PatchRecordsResult {
	return similapi.PatchRecordsResult{
		Deleted:  int(prr.Deleted),
		Upserted: int(prr.Upserted),
	}
}

func searchRecordsResultItems2Rest(srr *index.SearchRecordsResultItem) similapi.SearchRecord {
	return similapi.SearchRecord{
		IndexRecord:     record2Rest(srr.IndexRecord),
		IndexId:         srr.IndexId,
		Score:           cast.Value(srr.Score, -1.0),
		MatchedKeywords: srr.MatchedKeywords,
	}
}

func searchRecordsResult2Rest(srr *indrecord2Restex.SearchRecordsResult) similapi.SearchResult {
	res := similapi.SearchResult{}
	if srr == nil {
		return res
	}
	res.Records = make([]similapi.SearchRecord, len(srr.Items))
	for i, sr := range srr.Items {
		res.Records[i] = searchRecordsResultItems2Rest(sr)
	}
	res.NextPageId = srr.NextPageId
	res.Total = int(srr.Total)
	return res
}

func searchRequest2Proto(sr similapi.SearchRequest) *index.SearchRecordsRequest {
	return &index.SearchRecordsRequest{
		Text:         sr.Text,
		Tags:         sr.Tags,
		Distinct:     cast.Ptr(sr.Distinct),
		Limit:        cast.Ptr(int64(sr.Limit)),
		PageId:       cast.Ptr(sr.PageId),
		OrderByScore: cast.Ptr(sr.OrderByScore),
		IndexIDs:     sr.IndexIDs,
		Offset:       cast.Ptr(int64(sr.Offset)),
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
		Id:      r.Id,
		Segment: r.Segment,
		Vector:  r.Vector,
	}
}

func listRecordsResult2Rest(lrr *index.ListRecordsResult) similapi.RecordsResult {
	if lrr == nil {
		return similapi.RecordsResult{}
	}
	return similapi.RecordsResult{
		Records:    records2Rest(lrr.Records),
		NextPageId: lrr.NextRecordId,
		Total:      int(lrr.Total),
	}
}

func index2Rest(r *index.Index) similapi.Index {
	if r == nil {
		return similapi.Index{}
	}
	return similapi.Index{
		Id:        r.Id,
		Format:    r.Format,
		Tags:      r.Tags,
		CreatedAt: protoTime2Time(r.CreatedAt),
	}
}

func indexes2Rest(is *index.Indexes) similapi.Indexes {
	res := similapi.Indexes{}
	if is == nil {
		return res
	}
	res.NextPageId = is.NextIndexId
	res.Total = int(is.Total)
	res.Indexes = make([]similapi.Index, len(is.Indexes))
	for i, idx := range is.Indexes {
		res.Indexes[i] = index2Rest(idx)
	}
	return res
}

func index2Proto(i similapi.Index) *index.Index {
	return &index.Index{
		Id:        i.Id,
		Format:    i.Format,
		Tags:      i.Tags,
		CreatedAt: timestamppb.New(i.CreatedAt),
	}
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

func protoTime2Time(pt *timestamppb.Timestamp) time.Time {
	if pt == nil {
		return time.Time{}
	}
	return pt.AsTime()
}
