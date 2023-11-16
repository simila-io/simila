package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/cast"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/simila-io/simila/api/gen/index/v1"
	"sort"
	"strings"
)

type (
	// Commands represents all known commands
	Commands struct {
		cmds []Command
		ctx  context.Context
		isc  index.ServiceClient
	}

	Command interface {
		Run(prompt string) error
		Prefix() string

		shortDescription() string
		description() string
	}

	cmdHelp struct {
		cs *Commands
	}

	cmdListIndexes struct {
		cs *Commands
	}

	cmdSearch struct {
		cs *Commands
	}
)

const (
	Spaces = " \t\n\v\f\r\x85\xA0"
)

func New(ctx context.Context, isc index.ServiceClient) *Commands {
	cs := &Commands{ctx: ctx, isc: isc}
	cs.cmds = append(cs.cmds, cmdHelp{cs: cs})
	cs.cmds = append(cs.cmds, cmdListIndexes{cs: cs})
	cs.cmds = append(cs.cmds, cmdSearch{cs: cs})
	return cs
}

// returns list of all known commands
func (cs *Commands) ListCommandsNames() []string {
	res := []string{}
	for _, c := range cs.cmds {
		res = append(res, c.Prefix())
	}
	return res
}

// returns a command by its name
func (cs *Commands) GetCommand(prompt string) (Command, error) {
	prompt = strings.TrimLeft(prompt, Spaces)
	for _, c := range cs.cmds {
		if strings.HasPrefix(prompt, c.Prefix()) {
			return c, nil
		}
	}
	cmd := strings.SplitN(prompt, Spaces, 2)
	return nil, fmt.Errorf("unknown command %s ", cmd[0])
}

// -------------------------------- cmdHelp ----------------------------------
func (c cmdHelp) Run(prompt string) error {
	for _, c := range c.cs.cmds {
		if strings.HasPrefix(prompt, c.Prefix()) {
			fmt.Println(c.description())
			return nil
		}
	}
	fmt.Print(c.description())
	return nil
}

func (c cmdHelp) shortDescription() string {
	return "help <cmd> - prints help or a description by the command provided"
}

func (c cmdHelp) description() string {
	s := []string{}
	for _, c := range c.cs.cmds {
		s = append(s, c.shortDescription())
	}
	sort.Strings(s)
	var sb strings.Builder
	for _, line := range s {
		sb.WriteString(fmt.Sprintln(line))
	}
	return sb.String()
}

func (c cmdHelp) Prefix() string {
	return "help"
}

// -------------------------------- cmdListIndexes ---------------------------------
func (c cmdListIndexes) Run(prompt string) error {
	req := &index.ListRequest{}
	var asTable bool
	params := parseParams(prompt)
	for k, v := range params {
		switch k {
		case "startIndex":
			req.StartIndexId = strings.Trim(v, Spaces)
		case "tags":
			tags := map[string]string{}
			if err := json.Unmarshal(cast.StringToByteArray(v), &tags); err != nil {
				return fmt.Errorf("the tags value %q is wrong. Expecting tags to be a json map", v)
			}
			req.Tags = tags
		case "format":
			req.Format = cast.Ptr(strings.Trim(v, Spaces))
		case "limit":
			var limit int
			if err := json.Unmarshal(cast.StringToByteArray(v), &limit); err != nil || limit <= 0 {
				return fmt.Errorf("the limit value %s is wrong. limit must be a positive number", v)
			}
			req.Limit = cast.Ptr(int64(limit))
		case "as-table":
			if err := json.Unmarshal(cast.StringToByteArray(v), &asTable); err != nil {
				return fmt.Errorf("the as-table value %s is wrong. It must be a boolean value true/false", v)
			}
		default:
			return fmt.Errorf("unexpected parameter %s", k)
		}
	}

	indexes, err := c.cs.isc.List(c.cs.ctx, req)
	if err != nil {
		return err
	}
	if asTable {
		c.printAsTable(indexes)
	} else {
		b, _ := json.MarshalIndent(indexes, "", "  ")
		fmt.Println(string(b))
	}
	return nil
}

func (c cmdListIndexes) printAsTable(idx *index.Indexes) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Id", "Format", "Tags", "CreatedAt")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, r := range idx.Indexes {
		tags := "{}"
		if len(r.Tags) > 0 {
			b, _ := json.Marshal(r.Tags)
			tags = string(b)
		}
		tbl.AddRow(
			r.Id,
			r.Format,
			tags,
			r.CreatedAt.AsTime(),
		)
	}

	tbl.Print()
	if idx.NextIndexId != nil && *idx.NextIndexId != "" {
		fmt.Println("NextIndexId: ", *idx.NextIndexId)
	}
	fmt.Println("Total: ", idx.Total)
}

func (c cmdListIndexes) shortDescription() string {
	return "list indexes <params> - allows to request index list"
}

func (c cmdListIndexes) description() string {
	return `
list indexes <params> - lists the known indexes. It accepts the following params:

	startIndex=<string> - the first index in the result
	tags={"a":"a", "b":"b"} - the indexes with the tags values
	format=<string> - the indexes for the format
	limit=<int> - the number of records in the response
	as-table=<bool> - prints the result in a table for
`
}

func (c cmdListIndexes) Prefix() string {
	return "list indexes"
}

// -------------------------------- cmdSearch ---------------------------------
func (c cmdSearch) Run(prompt string) error {
	req := &index.SearchRecordsRequest{}
	var asTable bool
	params := parseParams(prompt)
	for k, v := range params {
		switch k {
		case "text":
			req.Text = strings.Trim(v, Spaces)
		case "tags":
			tags := map[string]string{}
			if err := json.Unmarshal(cast.StringToByteArray(v), &tags); err != nil {
				return fmt.Errorf("the tags value %q is wrong. Expecting tags to be a json map", v)
			}
			req.Tags = tags
		case "indexes":
			var idxs []string
			if err := json.Unmarshal(cast.StringToByteArray(v), &idxs); err != nil {
				return fmt.Errorf("the list of indexes %q is wrong. Expecting a json list of strings", v)
			}
			req.IndexIDs = idxs
		case "limit":
			var limit int
			if err := json.Unmarshal(cast.StringToByteArray(v), &limit); err != nil || limit <= 0 {
				return fmt.Errorf("the limit value %s is wrong. limit must be a positive number", v)
			}
			req.Limit = cast.Ptr(int64(limit))
		case "distinct":
			var dist bool
			if err := json.Unmarshal(cast.StringToByteArray(v), &dist); err != nil {
				return fmt.Errorf("the distinct value %s is wrong. distinct must be a boolean value true/false", v)
			}
			req.Distinct = cast.Ptr(dist)
		case "as-table":
			if err := json.Unmarshal(cast.StringToByteArray(v), &asTable); err != nil {
				return fmt.Errorf("the as-table value %s is wrong. It must be a boolean value true/false", v)
			}
		default:
			return fmt.Errorf("unexpected parameter %s", k)
		}
	}
	req.OrderByScore = cast.Ptr(true)
	result, err := c.cs.isc.SearchRecords(c.cs.ctx, req)
	if err != nil {
		return err
	}
	if asTable {
		c.printAsTable(result)
	} else {
		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
	}
	return nil
}

func (c cmdSearch) printAsTable(srr *index.SearchRecordsResult) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	//columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Score", "IdxId", "Keywords", "Segment")
	tbl.WithHeaderFormatter(headerFmt)

	for _, r := range srr.Items {
		tbl.AddRow(
			cutStr(fmt.Sprintf("%.2f", cast.Value(r.Score, -1.0)), 5),
			cutStr(r.IndexId, 16),
			cutStr(fmt.Sprintf("%v", r.MatchedKeywords), 40),
			cutToKeyword(r.IndexRecord.Segment, r.MatchedKeywords[0], 80),
		)
	}

	tbl.Print()
	fmt.Println("Total: ", srr.Total)
}

func cutToKeyword(s, kw string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	idx := strings.Index(s, kw)
	if idx > maxLen && maxLen > 13 {
		s = "..." + s[maxLen-10:]
	}
	return cutStr(s, maxLen)
}

func cutStr(s string, maxLen int) string {
	if len(s) < maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (c cmdSearch) shortDescription() string {
	return "search <params> - run the search request across known indexes"
}

func (c cmdSearch) description() string {
	return `
search <params> - returns the search results. It accepts the following params:

	text=<string> - the query text
	tags={"a":"a", "b":"b"} - the indexes with the tags values
	indexes=["index1", "index2"] - the list of indexes to run the search through
	distinct=<bool> - one record per index in the result
	limit=<int> - the number of records in the response
	as-table=<bool> - prints the result in a table form
`
}

func (c cmdSearch) Prefix() string {
	return "search"
}

func parseParams(s string) map[string]string {
	vals := splitParams(s)
	res := map[string]string{}
	key := ""
	for i, v := range vals {
		if i == 0 {
			key = v
			continue
		}
		k := key
		if i < len(vals)-1 {
			parts := strings.Split(v, " ")
			if len(parts) > 0 {
				key = parts[len(parts)-1]
				v = strings.TrimRight(v[:len(v)-len(key)], Spaces)
			}
		}
		res[k] = unquote(v)
	}
	return res
}

func splitParams(s string) []string {
	ss := strings.Split(s, "=")
	res := []string{}
	var sb strings.Builder
	for _, v := range ss {
		if len(v) > 0 && v[len(v)-1] == '\\' {
			sb.WriteString(v[:len(v)-1])
			sb.WriteString("=")
			continue
		}
		sb.WriteString(v)
		res = append(res, strings.Trim(sb.String(), Spaces))
		sb.Reset()
	}
	if sb.Len() > 0 {
		res = append(res, sb.String())
	}
	return res
}

func unquote(s string) string {
	s = strings.Trim(s, " ")
	if len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
