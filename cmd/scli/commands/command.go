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

	cmdListRecords struct {
		cs *Commands
	}

	cmdSearch struct {
		cs *Commands
	}

	cmdListNodes struct {
		cs *Commands
	}
)

const (
	Spaces = " \t\n\v\f\r\x85\xA0"
)

func New(ctx context.Context, isc index.ServiceClient) *Commands {
	cs := &Commands{ctx: ctx, isc: isc}
	cs.cmds = append(cs.cmds, cmdHelp{cs: cs})
	cs.cmds = append(cs.cmds, cmdListRecords{cs: cs})
	cs.cmds = append(cs.cmds, cmdSearch{cs: cs})
	cs.cmds = append(cs.cmds, cmdListNodes{cs: cs})
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

// -------------------------------- cmdListRecords ---------------------------------
func (c cmdListRecords) Run(prompt string) error {
	req := &index.ListRequest{}
	var asTable bool
	params := parseParams(prompt)

	for k, v := range params {
		switch k {
		case "path":
			req.Path = strings.Trim(v, Spaces)
		case "format":
			req.Format = cast.Ptr(strings.Trim(v, Spaces))
		case "limit":
			var limit int
			if err := json.Unmarshal(cast.StringToByteArray(v), &limit); err != nil || limit <= 0 {
				return fmt.Errorf("the limit value %s is wrong. limit must be a positive number", v)
			}
			req.Limit = cast.Ptr(int64(limit))
		case "page-id":
			req.PageId = cast.Ptr(strings.Trim(v, Spaces))
		case "as-table":
			if err := json.Unmarshal(cast.StringToByteArray(v), &asTable); err != nil {
				return fmt.Errorf("the as-table value %s is wrong. It must be a boolean value true/false", v)
			}
		default:
			return fmt.Errorf("unexpected parameter %s", k)
		}
	}

	records, err := c.cs.isc.ListRecords(c.cs.ctx, req)
	if err != nil {
		return err
	}
	if asTable {
		c.printAsTable(records)
	} else {
		b, _ := json.MarshalIndent(records, "", "  ")
		fmt.Println(string(b))
	}
	return nil
}

func (c cmdListRecords) printAsTable(lrr *index.ListRecordsResult) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Id", "Format", "RM", "Segment", "Vector")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, r := range lrr.Records {
		tbl.AddRow(
			r.Id,
			r.Format,
			r.RankMultiplier,
			r.Segment,
			r.Vector,
		)
	}

	tbl.Print()
	if lrr.NextPageId != nil && *lrr.NextPageId != "" {
		fmt.Println("NextPageId: ", *lrr.NextPageId)
	}
	fmt.Println("Total: ", lrr.Total)
}

func (c cmdListRecords) shortDescription() string {
	return "list records <params> - allows to request index records for a path"
}

func (c cmdListRecords) description() string {
	return `
list records <params> - lists the known index records for a path. It accepts the following params:

	path=<string> - FQNP - fully qualified node path
	format=<string> - the indexes for the format
	limit=<int> - the number of records in the response
	as-table=<bool> - prints the result in a table 
	page-id=<string> - next page if needed
`
}

func (c cmdListRecords) Prefix() string {
	return "list records"
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
		case "path":
			req.Path = strings.Trim(v, Spaces)
		case "limit":
			var limit int
			if err := json.Unmarshal(cast.StringToByteArray(v), &limit); err != nil || limit <= 0 {
				return fmt.Errorf("the limit value %s is wrong. limit must be a positive number", v)
			}
			req.Limit = cast.Ptr(int64(limit))
		case "strict":
			var dist bool
			if err := json.Unmarshal(cast.StringToByteArray(v), &dist); err != nil {
				return fmt.Errorf("the strict value %s is wrong. distinct must be a boolean value true/false", v)
			}
			req.Strict = cast.Ptr(dist)
		case "as-table":
			if err := json.Unmarshal(cast.StringToByteArray(v), &asTable); err != nil {
				return fmt.Errorf("the as-table value %s is wrong. It must be a boolean value true/false", v)
			}
		default:
			return fmt.Errorf("unexpected parameter %s", k)
		}
	}
	result, err := c.cs.isc.Search(c.cs.ctx, req)
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

	tbl := table.New("Score", "Path", "Keywords", "Segment")
	tbl.WithHeaderFormatter(headerFmt)

	for _, r := range srr.Items {
		tbl.AddRow(
			cutStr(fmt.Sprintf("%.2f", cast.Value(r.Score, -1.0)), 5),
			cutStr(r.Path, 16),
			cutStr(fmt.Sprintf("%v", r.MatchedKeywords), 40),
			cutToKeyword(r.Record.Segment, r.MatchedKeywords[0], 80),
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
	return "search <params> - run the search request across known index records"
}

func (c cmdSearch) description() string {
	return `
search <params> - returns the search results. It accepts the following params:

	text=<string> - the query text
	tags={"a":"a", "b":"b"} - the indexes with the tags values
	path=<string> - FQNP the path to the node to run the search for
	strict=<bool> - run the search for the node only (excluding its children, if any)
	limit=<int> - the number of records in the response
	as-table=<bool> - prints the result in a table form
`
}

func (c cmdSearch) Prefix() string {
	return "search"
}

// -------------------------------- cmdListNodes ----------------------------------
func (c cmdListNodes) Run(prompt string) error {
	nodes, err := c.cs.isc.ListNodes(c.cs.ctx, &index.Path{Path: prompt})
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(nodes, "", "  ")
	fmt.Println(string(b))
	return nil
}

func (c cmdListNodes) shortDescription() string {
	return "ls <path> - list nodes for the path"
}

func (c cmdListNodes) description() string {
	return `
ls <path> - prints all nodes for the path. The path is a FQNP to the node, whose children nodes should be printed
`
}

func (c cmdListNodes) Prefix() string {
	return "ls"
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
