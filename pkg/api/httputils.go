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
	"github.com/acquirecloud/golibs/cast"
	"github.com/simila-io/simila/api/gen/index/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type (
	createIndexRequest struct {
		// id contains the index identifier. It may be generated or provided. If provided, caller must
		// support it. id cannot be more than 256 bytes long
		Id string `json:"id,omitempty"`
		// format name. Format must exist
		Format string `json:"format,omitempty"`
		// tags associated with the index. May be empty.
		Tags map[string]string `json:"tags,omitempty"`
		// document contains the binary data for the format provided. It may be empty
		Document []byte `json:"document,omitempty"`
		// records contains the list of records that can be added to the index when it is created
		Records []*record `json:"records,omitempty"`
	}

	record struct {
		// id is the record id - this field is populated by parser or it is provided when records are created.
		Id string `json:"id,omitempty"`
		// segment contains the searchable text for the record
		Segment string `json:"segment,omitempty"`
		// vector is the list of the byte values (stringified) ordered according the basis fields definition
		Vector json.RawMessage `json:"vector,omitempty"`
	}

	patchRecordsRequest struct {
		// id is the patched index id
		Id string `json:"id,omitempty"`
		// upsertRecords contains the list of records that should be inserted or updated
		UpsertRecords []*record `json:"upsertRecords,omitempty"`
		// deleteRecords contains the list of records that should be deleted
		DeleteRecords []*record `json:"deleteRecords,omitempty"`
	}

	searchRecordsResult struct {
		Records    []*searchRecord `json:"records,omitempty"`
		NextPageId *string         `json:"nextPageId,omitempty"`
		Total      int             `json:"total"`
	}

	searchRecord struct {
		IndexId         string   `json:"indexId,omitempty"`
		IndexRecord     *record  `json:"indexRecord,omitempty"`
		MatchedKeywords []string `json:"matchedKeywords,omitempty"`
		Score           *int     `json:"score,omitempty"`
	}

	listRecordsResult struct {
		Records      []*record `json:"records,omitempty"`
		NextRecordId *string   `json:"nextRecordId,omitempty"`
		Total        int       `json:"total"`
	}

	indexStruct struct {
		// id the index uniquely identifier
		Id string `json:"id,omitempty"`
		// format - the index format
		Format string `json:"format,omitempty"`
		// tags the list of key:value pairs associated with the index
		Tags map[string]string `json:"tags,omitempty"`
		// createdAt the timestamp when the index was created
		CreatedAt time.Time `json:"createdAt,omitempty"`
	}
)

func createIndexRequest2Proto(ci createIndexRequest) *index.CreateIndexRequest {
	return &index.CreateIndexRequest{
		Id:       ci.Id,
		Format:   ci.Format,
		Tags:     ci.Tags,
		Document: ci.Document,
		Records:  records2Proto(ci.Records),
	}
}

func records2Proto(rs []*record) []*index.Record {
	if len(rs) == 0 {
		return nil
	}
	res := make([]*index.Record, len(rs))
	for i, r := range rs {
		res[i] = record2Proto(r)
	}
	return res
}

func record2Proto(r *record) *index.Record {
	return &index.Record{
		Id:      r.Id,
		Segment: r.Segment,
		Vector:  r.Vector,
	}
}

func patchIndexRecordsRequest2Proto(pr patchRecordsRequest) *index.PatchRecordsRequest {
	return &index.PatchRecordsRequest{
		Id:            pr.Id,
		UpsertRecords: records2Proto(pr.UpsertRecords),
		DeleteRecords: records2Proto(pr.DeleteRecords),
	}
}

func searchRecordsResult2Rest(srr *index.SearchRecordsResult) *searchRecordsResult {
	if srr == nil {
		return nil
	}
	return &searchRecordsResult{
		Records:    searchRecordsResultItems2Rest(srr.Items),
		NextPageId: srr.NextPageId,
		Total:      int(srr.Total),
	}
}

func searchRecordsResultItems2Rest(sri []*index.SearchRecordsResultItem) []*searchRecord {
	if len(sri) == 0 {
		return nil
	}
	res := make([]*searchRecord, len(sri))
	for i, r := range sri {
		res[i] = searchRecordsResultItem2Rest(r)
	}
	return res
}

func searchRecordsResultItem2Rest(s *index.SearchRecordsResultItem) *searchRecord {
	sri := &searchRecord{
		IndexId:         s.IndexId,
		IndexRecord:     record2Rest(s.IndexRecord),
		MatchedKeywords: s.MatchedKeywords,
	}
	if s.Score != nil {
		sri.Score = cast.Ptr(int(*s.Score))
	}
	return sri
}

func records2Rest(rs []*index.Record) []*record {
	if len(rs) == 0 {
		return nil
	}
	res := make([]*record, len(rs))
	for i, r := range rs {
		res[i] = record2Rest(r)
	}
	return res
}

func record2Rest(r *index.Record) *record {
	return &record{
		Id:      r.Id,
		Segment: r.Segment,
		Vector:  r.Vector,
	}
}

func listRecordsResult2Proto(lrr *index.ListRecordsResult) *listRecordsResult {
	if lrr == nil {
		return nil
	}
	return &listRecordsResult{
		Records:      records2Rest(lrr.Records),
		NextRecordId: lrr.NextRecordId,
		Total:        int(lrr.Total),
	}
}

func index2Rest(r *index.Index) *indexStruct {
	if r == nil {
		return nil
	}
	return &indexStruct{
		Id:        r.Id,
		Format:    r.Format,
		Tags:      r.Tags,
		CreatedAt: protoTime2Time(r.CreatedAt),
	}
}

func protoTime2Time(pt *timestamppb.Timestamp) time.Time {
	if pt == nil {
		return time.Time{}
	}
	return pt.AsTime()
}
