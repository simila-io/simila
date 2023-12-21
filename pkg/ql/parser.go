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
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

type (
	// Expression is an AST element which describes a series of OR conditions
	Expression struct {
		Or []*OrCondition `@@ { "OR" @@ }`
	}

	// OrCondition is an AST element which describes a series of AND conditions
	OrCondition struct {
		And []*XCondition `@@ { "AND" @@ }`
	}

	// XCondition is an AST element which groups either a Condition object or an Expression object
	XCondition struct {
		Not  bool        ` [@"NOT"] `
		Cond *Condition  `( @@`
		Expr *Expression `| "(" @@ ")")`
	}

	// Condition is a unary or binary logical operation which has First mandatory param and
	// optional operation and second param (optional)
	Condition struct {
		FirstParam  Param  `  @@`
		Op          string ` {@("<"|">"|">="|"<="|"!="|"="|"IN"|"LIKE")`
		SecondParam *Param ` @@}`
	}

	// Param describes a parameter either a constant (string or number), function, identifier or an array of constants
	Param struct {
		Const      *Const    ` @@`
		Function   *Function ` | @@`
		Identifier string    ` | @Ident`
		Array      []*Const  `|"[" (@@ {"," @@})?"]"`
	}

	// Const contains the constant either string or float32 value
	Const struct {
		Number float32 ` @Number`
		String string  ` | @String`
	}

	// Function is a functional parameter
	Function struct {
		Name   string   ` @Ident `
		Params []*Param ` "(" (@@ {"," @@})? ")"`
	}

	// Dialect describes how the parameter of specific type or name should be treated. Dialect has a name,
	// either predefined for String(StringParamID), Number(NumberParamID) or the Array(ArrayParamID) or
	// it is defined for the specific identifier or function name.
	Dialect struct {
		// Flags defines how the parameter can be treated. Is it Unary or binary operation, can the parameter
		// be a lvalue etc.
		Flags int
		// Translate is the function (can be nil), which is called for translation the parameter p to the dialect. For example,
		// on the ql level the function hasPrefix(path, "abc") may be translated to the Postgres SQL:
		// position('abc' in record.path) = 1
		Translate func(tr Translator, sb *strings.Builder, p Param) error
	}

	// Translator struct allows to turn AST objects (Expression, Condition etc.) to
	// the SQL statements according to the dialect provided
	Translator struct {
		dialects map[string]Dialect
	}
)

var (
	sqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Keyword`, `(?i)\b(AND|OR|NOT|IN|LIKE)\b`},
		{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{`String`, `'[^']*'|"[^"]*"`},
		{`Operators`, `!=|<=|>=|[,()=<>\]\[]`},
		{"whitespace", `\s+`},
	})

	parser = participle.MustBuild[Expression](
		participle.Lexer(sqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)

	// PqFilterConditionsDialect is a set of specific dialects for
	// translating filter conditions into Postgres where condition.
	PqFilterConditionsDialect = map[string]Dialect{
		StringParamID: {
			Flags: PfRValue | PfComparable, // strings are rvalues only
			Translate: func(tr Translator, sb *strings.Builder, p Param) error {
				// use single quotes for string constants
				sb.WriteString("'")
				sb.WriteString(p.Const.String)
				sb.WriteString("'")
				return nil
			},
		},
		NumberParamID: {Flags: PfRValue | PfComparable}, // numbers are rvalues only
		ArrayParamID:  {Flags: PfRValue},                // arrays are rvalues only

		// path identifier, maybe a part of operations like `path = "/org1/folders1/doc1.txt"` etc.
		"path": {
			Flags: PfLValue | PfComparable | PfInLike,
			Translate: func(tr Translator, sb *strings.Builder, p Param) error {
				sb.WriteString("n.path")
				return nil
			},
		},

		"node": {
			Flags: PfLValue | PfComparable,
			Translate: func(tr Translator, sb *strings.Builder, p Param) error {
				sb.WriteString("concat(n.path, n.name)")
				return nil
			},
		},

		// format identifier is used as LValue of comparable expressions
		"format": {
			Flags: PfLValue | PfComparable | PfInLike,
			Translate: func(tr Translator, sb *strings.Builder, p Param) error {
				sb.WriteString("ir.format")
				return nil
			},
		},

		// tag function is written the way -> 'tag("abc") in ["1", "2", "3"]' or 'tag("t1") = "aaa"'
		"tag": {
			Flags: PfLValue | PfComparable | PfRValue | PfInLike,
			Translate: func(tr Translator, sb *strings.Builder, p Param) error {
				if p.Function == nil {
					return fmt.Errorf("tag must be a function: %w", errors.ErrInvalid)
				}
				if len(p.Function.Params) != 1 {
					return fmt.Errorf("tag() function expects only one parameter - the name of the tag: %w", errors.ErrInvalid)
				}
				if p.Function.Params[0].id() != StringParamID {
					return fmt.Errorf("tag() function expects the tag name (string) as the parameter: %w", errors.ErrInvalid)
				}
				sb.WriteString("n.tags ->> ")
				_ = tr.Param2Sql(sb, p.Function.Params[0])
				return nil
			},
		},

		// prefix(s, p) returns whether the s has prefix p
		"prefix": {
			Flags: PfLValue | PfNop,
			Translate: func(tr Translator, sb *strings.Builder, p Param) error {
				if p.Function == nil {
					return fmt.Errorf("prefix must be a function: %w", errors.ErrInvalid)
				}
				if len(p.Function.Params) != 2 {
					return fmt.Errorf("prefix(s, p) function expects two parameters: %w", errors.ErrInvalid)
				}
				sb.WriteString(" position(")
				_ = tr.Param2Sql(sb, p.Function.Params[1])
				sb.WriteString(" in ")
				_ = tr.Param2Sql(sb, p.Function.Params[0])
				sb.WriteString(") = 1")
				return nil
			},
		},
	}
)

const (
	StringParamID = "__string__"
	NumberParamID = "__number__"
	ArrayParamID  = "__array__"

	// PfLValue the parameter can be a lvalue in the condition
	PfLValue = 1 << 0
	// PfRValue the parameter can be a rvalue in the condition
	PfRValue = 1 << 1
	// PfNop the parameter cannot have any operation
	PfNop = 1 << 2
	// PfComparable the parameter can be compared: <, >, !=, =, >=, <=
	PfComparable = 1 << 3
	// PfInLike the IN or LIKE operations are allowed for the param
	PfInLike = 1 << 4
)

// NewTranslator creates new Translator with dialects provided
func NewTranslator(dialects map[string]Dialect) Translator {
	return Translator{dialects: dialects}
}

// id returns the param id by its type:
// - string: StringParamID
// - number: NumberParamID
// - function: the function name
// - identifier: the identifier name
// - array: ArrayParamID
func (p Param) id() string {
	if p.Const != nil {
		if p.Const.String != "" {
			return StringParamID
		}
		return NumberParamID
	}
	if p.Function != nil {
		return p.Function.Name
	}
	if p.Identifier != "" {
		return p.Identifier
	}
	return ArrayParamID
}

// Name returns "value" of the constants (strings, numbers and the arrays) and names for the functions and identifiers
func (p Param) Name(full bool) string {
	if p.Const != nil {
		return p.Const.Value()
	}
	if p.Function != nil {
		return p.Function.Name
	}
	if p.Identifier != "" {
		return p.Identifier
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, c := range p.Array {
		if i > 0 {
			sb.WriteString(", ")
		}
		if !full && i > 3 && len(p.Array) > 10 {
			sb.WriteString(fmt.Sprintf("... and %d more", len(p.Array)-4))
			break
		}
		sb.WriteString(c.Value())
	}
	sb.WriteString("]")
	return sb.String()
}

// Value returns string value of the constant
func (c Const) Value() string {
	if c.String != "" {
		return fmt.Sprintf("%q", c.String)
	}
	return fmt.Sprintf("%f", c.Number)
}

// Translate translates the expression string to string according to the dialect of the translator
func (tr Translator) Translate(sb *strings.Builder, expr string) error {
	expr = strings.TrimSpace(expr)
	if len(expr) == 0 {
		return nil
	}
	e, err := parser.ParseString("", expr)
	if err != nil {
		return fmt.Errorf("failed to parse expression=%q: %w", expr, err)
	}
	if err = tr.Expression2Sql(sb, e); err != nil {
		return fmt.Errorf("failed to translate expression=%q: %w", expr, err)
	}
	return nil
}

// Expression2Sql turns the AST object e to the query string according to the dialect of the translator
func (tr Translator) Expression2Sql(sb *strings.Builder, e *Expression) error {
	for i, oc := range e.Or {
		if i > 0 {
			sb.WriteString(" OR ")
		}
		if err := tr.OrCondition2Sql(sb, oc); err != nil {
			return err
		}
	}
	return nil
}

// OrCondition2Sql turns the AST object oc to the query string according to the dialect of the translator
func (tr Translator) OrCondition2Sql(sb *strings.Builder, oc *OrCondition) error {
	for i, xc := range oc.And {
		if i > 0 {
			sb.WriteString(" AND ")
		}
		if err := tr.XCondition2Sql(sb, xc); err != nil {
			return err
		}
	}
	return nil
}

// XCondition2Sql turns the AST object xc to the query string according to the dialect of the translator
func (tr Translator) XCondition2Sql(sb *strings.Builder, xc *XCondition) error {
	if xc.Not {
		sb.WriteString(" NOT ")
	}
	if xc.Expr != nil {
		sb.WriteString("(")
		defer sb.WriteString(")")
		return tr.Expression2Sql(sb, xc.Expr)
	}
	return tr.Condition2Sql(sb, xc.Cond)
}

// Condition2Sql turns the AST object c to the query string according to the dialect of the translator
func (tr Translator) Condition2Sql(sb *strings.Builder, c *Condition) error {
	d, ok := tr.dialects[c.FirstParam.id()]
	if !ok {
		return fmt.Errorf("unknown parameter %s: %w", c.FirstParam.Name(false), errors.ErrInvalid)
	}
	if d.Flags&PfLValue == 0 {
		return fmt.Errorf("parameter %s cannot be on the left side of the condition: %w", c.FirstParam.Name(false), errors.ErrInvalid)
	}
	if c.Op == "" {
		if d.Flags&PfNop == 0 {
			return fmt.Errorf("parameter %s should be compared with something in a condition: %w", c.FirstParam.Name(false), errors.ErrInvalid)
		}
		return tr.Param2Sql(sb, &c.FirstParam)
	}
	if d.Flags&PfNop != 0 {
		return fmt.Errorf("parameter %s cannot be compared (%s) in the condition: %w", c.FirstParam.Name(false), c.Op, errors.ErrInvalid)
	}

	if c.SecondParam == nil {
		return fmt.Errorf("wrong condition for the param %s and the operation %q - no second parameter: %w", c.FirstParam.Name(false), c.Op, errors.ErrInvalid)
	}
	d2, ok := tr.dialects[c.SecondParam.id()]
	if !ok {
		return fmt.Errorf("unknown second parameter %s: %w", c.SecondParam.Name(false), errors.ErrInvalid)
	}
	if d2.Flags&PfRValue == 0 {
		return fmt.Errorf("parameter %s cannot be on the right side of the condition: %w", c.SecondParam.Name(false), errors.ErrInvalid)
	}
	if d2.Flags&PfNop != 0 {
		return fmt.Errorf("parameter %s cannot be compared (%s) in the condition: %w", c.SecondParam.Name(false), c.Op, errors.ErrInvalid)
	}

	op := strings.ToUpper(c.Op)
	switch op {
	case "<", ">", "<=", ">=", "!=", "=":
		if d.Flags&PfComparable == 0 {
			return fmt.Errorf("the first parameter %s is not applicable for the operation %s: %w", c.FirstParam.Name(false), c.Op, errors.ErrInvalid)
		}
		if d2.Flags&PfComparable == 0 {
			return fmt.Errorf("the first parameter %s is not applicable for the operation %s: %w", c.SecondParam.Name(false), c.Op, errors.ErrInvalid)
		}
	case "IN":
		if d.Flags&PfInLike == 0 {
			return fmt.Errorf("the first parameter %s is not applicable for the IN : %w", c.FirstParam.Name(false), errors.ErrInvalid)
		}
		if c.SecondParam.id() != ArrayParamID {
			return fmt.Errorf("the second parameter %s must be an array: %w", c.SecondParam.Name(false), errors.ErrInvalid)
		}
	case "LIKE":
		if d.Flags&PfInLike == 0 {
			return fmt.Errorf("the first parameter %s is not applicable for the LIKE : %w", c.FirstParam.Name(false), errors.ErrInvalid)
		}
		if c.SecondParam.id() != StringParamID {
			return fmt.Errorf("the right value(%s) of LIKE must be a string: %w", c.SecondParam.Name(false), errors.ErrInvalid)
		}
	default:
		return fmt.Errorf("unknown operation %s: %w", c.Op, errors.ErrInvalid)
	}
	if err := tr.Param2Sql(sb, &c.FirstParam); err != nil {
		return err
	}
	sb.WriteString(" ")
	sb.WriteString(op)
	sb.WriteString(" ")
	return tr.Param2Sql(sb, c.SecondParam)
}

// Param2Sql turns the AST object p to the query string according to the dialect of the translator
func (tr Translator) Param2Sql(sb *strings.Builder, p *Param) error {
	d, ok := tr.dialects[p.id()]
	if !ok {
		return fmt.Errorf("unknown parameter %s: %w", p.Name(false), errors.ErrInvalid)
	}
	if d.Translate != nil {
		return d.Translate(tr, sb, *p)
	}
	name := p.Name(true)
	sb.WriteString(name)
	if p.Function == nil {
		return nil
	}
	sb.WriteString("(")
	_ = tr.Params2Sql(sb, p.Function.Params)
	sb.WriteString(")")
	return nil
}

// Params2Sql turns the AST object prms to the query string according to the dialect of the translator
func (tr Translator) Params2Sql(sb *strings.Builder, prms []*Param) error {
	for i, prm := range prms {
		if i > 0 {
			sb.WriteString(", ")
		}
		if err := tr.Param2Sql(sb, prm); err != nil {
			return err
		}
	}
	return nil
}
