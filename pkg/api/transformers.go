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
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		MatchedKeywords: mItem.MatchedKeywords,
	}
	if includeScore {
		srr.Score = cast.Ptr(int64(cast.Value(&mItem.Score, 0)))
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
