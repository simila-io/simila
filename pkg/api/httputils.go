package api

import (
	"encoding/json"
	"github.com/simila-io/simila/api/gen/index/v1"
)

type (
	CreateIndexRequest struct {
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
		Records []*Record `json:"records,omitempty"`
	}

	Record struct {
		// id is the record id - this field is populated by parser or it is provided when records are created.
		Id string `json:"id,omitempty"`
		// segment contains the searchable text for the record
		Segment string `json:"segment,omitempty"`
		// vector is the list of the byte values (stringified) ordered according the basis fields definition
		Vector json.RawMessage `json:"vector,omitempty"`
	}

	PatchRecordsRequest struct {
		// id is the patched index id
		Id string `json:"id,omitempty"`
		// upsertRecords contains the list of records that should be inserted or updated
		UpsertRecords []*Record `json:"upsertRecords,omitempty"`
		// deleteRecords contains the list of records that should be deleted
		DeleteRecords []*Record `json:"deleteRecords,omitempty"`
	}

	SearchRecordsResult struct {
		Records    []*IndexRecord `json:"records,omitempty"`
		NextPageId *string        `json:"nextPageId,omitempty"`
	}

	IndexRecord struct {
		IndexId     string  `json:"indexId,omitempty"`
		IndexRecord *Record `json:"indexRecord,omitempty"`
	}

	ListRecordsResult struct {
		Records      []*Record `json:"records,omitempty"`
		NextRecordId *string   `json:"nextRecordId,omitempty"`
	}
)

func CreateIndexRequest2Proto(ci CreateIndexRequest) *index.CreateIndexRequest {
	return &index.CreateIndexRequest{
		Id:       ci.Id,
		Format:   ci.Format,
		Tags:     ci.Tags,
		Document: ci.Document,
		Records:  Records2Proto(ci.Records),
	}
}

func Records2Proto(rs []*Record) []*index.Record {
	if len(rs) == 0 {
		return nil
	}
	res := make([]*index.Record, len(rs))
	for i, r := range rs {
		res[i] = Record2Proto(r)
	}
	return res
}

func Record2Proto(r *Record) *index.Record {
	return &index.Record{
		Id:      r.Id,
		Segment: r.Segment,
		Vector:  r.Vector,
	}
}

func PatchIndexRecordsRequest2Proto(pr PatchRecordsRequest) *index.PatchRecordsRequest {
	return &index.PatchRecordsRequest{
		Id:            pr.Id,
		UpsertRecords: Records2Proto(pr.UpsertRecords),
		DeleteRecords: Records2Proto(pr.DeleteRecords),
	}
}

func SearchRecordsResult2Rest(srr *index.SearchRecordsResult) SearchRecordsResult {
	return SearchRecordsResult{
		Records:    IndexRecords2Rest(srr.Records),
		NextPageId: srr.NextPageId,
	}
}

func IndexRecords2Rest(irs []*index.IndexRecord) []*IndexRecord {
	if len(irs) == 0 {
		return nil
	}
	res := make([]*IndexRecord, len(irs))
	for i, r := range irs {
		res[i] = IndexRecord2Rest(r)
	}
	return res
}

func Records2Rest(rs []*index.Record) []*Record {
	if len(rs) == 0 {
		return nil
	}
	res := make([]*Record, len(rs))
	for i, r := range rs {
		res[i] = Record2Rest(r)
	}
	return res
}

func IndexRecord2Rest(ir *index.IndexRecord) *IndexRecord {
	return &IndexRecord{
		IndexId:     ir.IndexId,
		IndexRecord: Record2Rest(ir.IndexRecord),
	}
}

func Record2Rest(r *index.Record) *Record {
	return &Record{
		Id:      r.Id,
		Segment: r.Segment,
		Vector:  r.Vector,
	}
}

func ListRecordsResult2Proto(lrr *index.ListRecordsResult) ListRecordsResult {
	return ListRecordsResult{
		Records:      Records2Rest(lrr.Records),
		NextRecordId: lrr.NextRecordId,
	}
}
