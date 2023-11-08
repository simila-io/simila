package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/cast"
	"github.com/simila-io/simila/api/gen/index/v1"
	"sort"
	"strings"
)

type (
	// Commands represents all known commands
	Commands struct {
		cmds []Command
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
		ctx context.Context
		isc index.ServiceClient
	}

	cmdSearch struct {
		ctx context.Context
		isc index.ServiceClient
	}
)

const (
	Spaces = " \t\n\v\f\r\x85\xA0"
)

func New(ctx context.Context, isc index.ServiceClient) *Commands {
	cs := &Commands{}
	cs.cmds = append(cs.cmds, cmdHelp{cs: cs})
	cs.cmds = append(cs.cmds, cmdListIndexes{ctx: ctx, isc: isc})
	cs.cmds = append(cs.cmds, cmdSearch{ctx: ctx, isc: isc})
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
		default:
			return fmt.Errorf("unexpected parameter %s", k)
		}
	}

	indexes, err := c.isc.List(c.ctx, req)
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(indexes, "", "  ")
	fmt.Println(string(b))
	return nil
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
`
}

func (c cmdListIndexes) Prefix() string {
	return "list indexes"
}

// -------------------------------- cmdSearch ---------------------------------
func (c cmdSearch) Run(prompt string) error {
	req := &index.SearchRecordsRequest{}
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
		default:
			return fmt.Errorf("unexpected parameter %s", k)
		}
	}

	result, err := c.isc.SearchRecords(c.ctx, req)
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
	return nil
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
`
}

func (c cmdSearch) Prefix() string {
	return "search"
}

func parseParams(s string) map[string]string {
	vals := strings.Split(strings.Trim(s, Spaces), "=")
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
		res[k] = v
	}
	return res
}
