package ql

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type (
	Expression struct {
		Or []*OrCondition `@@ { "OR" @@ }`
	}

	OrCondition struct {
		And []*XCondition `@@ { "AND" @@ }`
	}

	XCondition struct {
		Not  bool        ` [@"NOT"] `
		Cond *Condition  `( @@`
		Expr *Expression `| "(" @@ ")")`
	}

	Condition struct {
		FirstParam  Param  `  @@`
		Op          string ` {@("<"|">"|">="|"<="|"!="|"="|"IN")`
		SecondParam *Param ` @@}`
	}

	Param struct {
		Const      *Const    ` @@`
		Function   *Function ` | @@`
		Identifier string    ` | @Ident`
		Array      []*Const  `|"[" (@@ {"," @@})?"]"`
	}

	Const struct {
		Number float32 ` @Number`
		String string  ` | @String`
	}

	Function struct {
		Name   string   ` @Ident `
		Params []*Param ` "(" (@@ {"," @@})? ")"`
	}
)

var (
	sqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Keyword`, `(?i)\b(AND|OR|NOT|IN)\b`},
		{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{`String`, `'[^']*'|"[^"]*"`},
		{`Operators`, `<>|!=|<=|>=|[,()=<>\]\[]`},
		{"whitespace", `\s+`},
	})

	parser = participle.MustBuild[Expression](
		participle.Lexer(sqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)
)
