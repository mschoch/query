//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package n1ql

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/logging"
)

func ParseStatement(input string) (algebra.Statement, error) {
	input = strings.TrimSpace(input)
	reader := strings.NewReader(input)
	lex := newLexer(NewLexer(reader))
	lex.parsingStmt = true
	lex.text = input
	doParse(lex)

	if len(lex.errs) > 0 {
		return nil, fmt.Errorf(strings.Join(lex.errs, " \n "))
	} else if lex.stmt == nil {
		return nil, fmt.Errorf("Input was not a statement.")
	} else {
		err := lex.stmt.Formalize()
		if err != nil {
			return nil, err
		} else {
			return lex.stmt, nil
		}
	}
}

func ParseExpression(input string) (expression.Expression, error) {
	input = strings.TrimSpace(input)
	reader := strings.NewReader(input)
	lex := newLexer(NewLexer(reader))
	doParse(lex)

	if len(lex.errs) > 0 {
		return nil, fmt.Errorf(strings.Join(lex.errs, " \n "))
	} else if lex.expr == nil {
		return nil, fmt.Errorf("Input was not an expression.")
	} else {
		return lex.expr, nil
	}
}

func doParse(lex *lexer) {
	defer func() {
		r := recover()
		if r != nil {
			lex.Error(fmt.Sprintf("Error while parsing: %v", r))

			// Log this error
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			logging.Errorf("Error while parsing: %v\n%s", r, string(buf[0:n]))
		}
	}()

	yyParse(lex)
}

type lexer struct {
	nex         *Lexer
	posParam    int
	errs        []string
	stmt        algebra.Statement
	expr        expression.Expression
	parsingStmt bool
	text        string
}

func newLexer(nex *Lexer) *lexer {
	return &lexer{
		nex:  nex,
		errs: make([]string, 0, 16),
	}
}

func (this *lexer) Lex(lval *yySymType) int {
	return this.nex.Lex(lval)
}

func (this *lexer) Error(s string) {
	if len(this.nex.stack) > 0 {
		s = s + " - at " + this.nex.Text()
	} else {
		s = s + " - at end of input"
	}

	this.errs = append(this.errs, s)
}

func (this *lexer) setStatement(stmt algebra.Statement) {
	this.stmt = stmt
}

func (this *lexer) setExpression(expr expression.Expression) {
	this.expr = expr
}

func (this *lexer) parsingStatement() bool { return this.parsingStmt }

func (this *lexer) getText() string { return this.text }

func (this *lexer) nextParam() int {
	this.posParam++
	return this.posParam
}
