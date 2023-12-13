package ql

import (
	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/assert"
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
