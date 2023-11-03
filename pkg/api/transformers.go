package api

import (
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v1"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toApiFormat(mFmt persistence.Format) *format.Format {
	return &format.Format{Name: mFmt.ID}
}

func toModelFormat(aFmt *format.Format) persistence.Format {
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
	return persistence.Index{ID: aIdx.Id, Format: aIdx.Format}
}

func toModelIndexFromApiCreateIdxReq(request *index.CreateIndexRequest) persistence.Index {
	return persistence.Index{ID: request.Id, Tags: request.Tags, Format: request.Format}
}

func toApiIndexRecord(mItem persistence.SearchQueryResultItem) *index.IndexRecord {
	return &index.IndexRecord{
		IndexId:     mItem.IndexID,
		IndexRecord: toApiRecord(mItem.IndexRecord),
	}
}

func toModelIndexRecord(aRec *index.IndexRecord) persistence.IndexRecord {
	return persistence.IndexRecord{
		ID:      aRec.IndexRecord.Id,
		IndexID: aRec.IndexId,
		Segment: aRec.IndexRecord.Segment,
		Vector:  aRec.IndexRecord.Vector,
	}
}

func toModelIndexRecordFromApiRecord(indexID string, aRec *index.Record) persistence.IndexRecord {
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
