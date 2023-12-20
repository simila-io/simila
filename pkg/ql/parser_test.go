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

package ql

import (
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParseParam(t *testing.T) {
	p := participle.MustBuild[Param](
		participle.Lexer(sqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)

	res, err := p.ParseString("", "1234")
	assert.Nil(t, err)
	assert.Equal(t, float32(1234.0), res.Const.Number)

	res, err = p.ParseString("", "'1234'")
	assert.Nil(t, err)
	assert.Equal(t, "1234", res.Const.String)

	res, err = p.ParseString("", "lala")
	assert.Nil(t, err)
	assert.Equal(t, "lala", res.Identifier)

	res, err = p.ParseString("", "lala ( )")
	assert.Nil(t, err)
	assert.Equal(t, Function{Name: "lala"}, *res.Function)

	res, err = p.ParseString("", "lala ( 1234)")
	assert.Nil(t, err)
	assert.Equal(t, Function{Name: "lala", Params: []*Param{{Const: &Const{Number: float32(1234)}}}}, *res.Function)

	_, err = p.ParseString("", "lala ( 1234,hhh)")
	assert.Nil(t, err)

	_, err = p.ParseString("", "[1234, 22]")
	assert.Nil(t, err)
}

func TestParseCondition(t *testing.T) {
	p := participle.MustBuild[Condition](
		participle.Lexer(sqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)

	c, err := p.ParseString("", "1234")
	assert.Nil(t, err)
	assert.Equal(t, float32(1234.0), c.FirstParam.Const.Number)
	assert.Nil(t, c.SecondParam)

	c, err = p.ParseString("", "f1() != f2('asdf')")
	assert.Nil(t, err)
	assert.Equal(t, "f1", c.FirstParam.Function.Name)
	assert.Equal(t, "!=", c.Op)
	assert.Equal(t, "f2", c.SecondParam.Function.Name)
}

func TestExpressions(t *testing.T) {
	testOk(t, "1234.34")
	testOk(t, "'string'")
	testOk(t, "path")
	testOk(t, "f()")
	testOk(t, "f(1)")
	testOk(t, "f(1, 2)")
	testOk(t, "1234 != 1234 and f()")
	testOk(t, "1234 != 1234 and (f(1234, path, f2(34, f1())) or path = 'sdf')")
	testOk(t, "tag('abc') in [1,2,3]")
}

func testOk(t *testing.T, e string) {
	_, err := parser.ParseString("", e)
	assert.Nil(t, err)
}

func TestCondition2Sql(t *testing.T) {
	dialects := map[string]Dialect{
		StringParamID: {Flags: PfRValue | PfComparable}, // strings are rvalues only
		NumberParamID: {Flags: PfRValue | PfComparable}, // numbers are rvalues only
		ArrayParamID:  {Flags: PfRValue},                // arrays are rvalues only
		"unary":       {Flags: PfLValue | PfNop},
		"binary":      {Flags: PfLValue | PfComparable},
		"inonly":      {Flags: PfLValue | PfInArray},
	}
	tr := NewTranslator(dialects)
	p := participle.MustBuild[Condition](
		participle.Lexer(sqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)
	c, err := p.ParseString("", "ptr = 123")
	assert.Nil(t, err)
	var sb strings.Builder
	assert.NotNil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "123 = \"abc\"")
	assert.Nil(t, err)
	assert.NotNil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "unary() = \"abc\"")
	assert.Nil(t, err)
	assert.NotNil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "binary()")
	assert.Nil(t, err)
	assert.NotNil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "binary() > unary")
	assert.Nil(t, err)
	assert.NotNil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "inonly() > 123")
	assert.Nil(t, err)
	assert.NotNil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "unary()")
	assert.Nil(t, err)
	assert.Nil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "binary() < 1234")
	assert.Nil(t, err)
	assert.Nil(t, tr.Condition2Sql(&sb, c))

	c, err = p.ParseString("", "inonly() in [1234, 3245]")
	assert.Nil(t, err)
	assert.Nil(t, tr.Condition2Sql(&sb, c))
}

func TestDialects(t *testing.T) {
	dialects := map[string]Dialect{
		StringParamID: {Flags: PfRValue | PfComparable, Translate: func(tr Translator, sb *strings.Builder, p Param) error {
			sb.WriteString("'")
			sb.WriteString(p.Const.String)
			sb.WriteString("'")
			return nil
		}},
		NumberParamID: {Flags: PfRValue | PfComparable}, // numbers are rvalues only
		ArrayParamID:  {Flags: PfRValue},                // arrays are rvalues only
		"unary": {Flags: PfLValue | PfNop, Translate: func(tr Translator, sb *strings.Builder, p Param) error {
			if p.Function == nil {
				return fmt.Errorf("unary is a function, not an identifier: %w", errors.ErrInvalid)
			}
			if len(p.Function.Params) != 1 {
				return fmt.Errorf("unary expects only one argument: %w", errors.ErrInvalid)
			}
			sb.WriteString("unary123(")
			if err := tr.Params2Sql(sb, p.Function.Params); err != nil {
				return err
			}
			sb.WriteString(")")
			return nil
		}},
		"binary": {Flags: PfLValue | PfComparable, Translate: func(tr Translator, sb *strings.Builder, p Param) error {
			sb.WriteString("table.param1")
			return nil
		}},
		"inonly": {Flags: PfLValue | PfInArray, Translate: func(tr Translator, sb *strings.Builder, p Param) error {
			sb.WriteString("table.param2")
			return nil
		}},
	}

	tr := NewTranslator(dialects)
	var sb strings.Builder
	e, err := parser.ParseString("", "unary(\"hello world\") and (binary != 234 or inonly in [1,2])")
	assert.Nil(t, err)
	assert.Nil(t, tr.Expression2Sql(&sb, e))
	assert.Equal(t, "unary123('hello world') AND (table.param1 != 234.000000 OR table.param2 IN [1.000000, 2.000000])", sb.String())
}

func TestPqFilterConditionsDialect(t *testing.T) {
	tr := NewTranslator(PqFilterConditionsDialect)

	var sb strings.Builder
	e, err := parser.ParseString("", "tag(1234) = \"234\"")
	assert.Nil(t, err)
	assert.NotNil(t, tr.Expression2Sql(&sb, e))

	e, err = parser.ParseString("", "tag('1234', 134) = \"234\"")
	assert.Nil(t, err)
	assert.NotNil(t, tr.Expression2Sql(&sb, e))

	e, err = parser.ParseString("", "tag('1234') = \"234\"")
	assert.Nil(t, err)
	assert.Nil(t, tr.Expression2Sql(&sb, e))

	sb.Reset()
	e, err = parser.ParseString("", "tag('abc') = tag(\"def\") and (prefix(path, \"/aaa/\") or format = 1234.3)")
	assert.Nil(t, err)
	assert.Nil(t, tr.Expression2Sql(&sb, e))
	assert.Equal(t, "n.tags ->> 'abc' = n.tags ->> 'def' AND ( position('/aaa/' in concat(n.path, n.name)) = 1 OR format = 1234.300049)", sb.String())
}
