//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

/*
Package expression provides expression evaluation for query and
indexing.
*/
package expression

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/couchbase/query/value"
)

/*
The type Expressions is defined as a slice of Expression. The
type CompositeExpressions is defined as a slice of Expressions.
*/
type Expressions []Expression
type CompositeExpressions []Expressions

/*
The Expression interface represents N1QL expressions.
*/
type Expression interface {
	fmt.Stringer
	json.Marshaler

	/*
	   It takes as input a Visitor type and returns an interface
	   and error. It represents a visitor pattern.
	*/
	Accept(visitor Visitor) (interface{}, error)

	/*
	   Type() returns the N1QL data type of the result of this
	   Expression. Type() allows you to infer the schema or shape
	   of query results before actually evaluating the query.
	*/
	Type() value.Type

	/*
	   Evaluate the expression for a given input and a particular
	   context.
	*/
	Evaluate(item value.Value, context Context) (value.Value, error)

	/*
	   Evaluate the expression for an indexing context. Support
	   multiple return values for array indexing.
	*/
	EvaluateForIndex(item value.Value, context Context) (value.Value, value.Values, error)

	/*
	   Value() returns the static / constant value of this
	   Expression, or nil. Expressions that depend on data,
	   clocks, or random numbers must return nil. Used in index
	   selection.
	*/
	Value() value.Value

	/*
	   Static() returns the static / constant equivalent of this
	   Expression, or nil. Expressions that depend on data,
	   clocks, or random numbers must return nil. Used in index
	   selection.
	*/
	Static() Expression

	/*
	   As per the N1QL specs this function returns the terminal
	   identifier in the case the expression is a path. It can
	   be thought of an expression alias. For example if for the
	   following select statement, b is the Alias. Select a.b.
	*/
	Alias() string

	/*
	   This method indicates if the expression can be used as a
	   secondary index key.
	*/
	Indexable() bool

	/*
	   True iff this Expression always returns MISSING if any of
	   its inputs are MISSING. This test is used in index
	   selection when an index contains the clause WHERE expr IS
	   NOT MISSING. False negatives are allowed.
	*/
	PropagatesMissing() bool

	/*
	   True iff this Expression always returns NULL if any of its
	   inputs is NULL. This test is used in index selection when
	   an index contains the clause WHERE expr IS NOT NULL or the
	   clause WHERE expr IS VALUED. False negatives are allowed.
	*/
	PropagatesNull() bool

	/*
	   Indicates if this expression is equivalent to the other
	   expression.  False negatives are allowed. Used in index
	   selection.
	*/
	EquivalentTo(other Expression) bool

	/*
	   Indicates if this expression depends on the other
	   expression.  False negatives are allowed. Used in index
	   selection.
	*/
	DependsOn(other Expression) bool

	/*
	   Indicates if this expression is covered by the list of
	   expressions; that is, this expression does not depend on
	   any stored data beyond the expressions.
	*/
	CoveredBy(exprs Expressions) bool

	/*
	   It is a utility function that returns the children of the
	   expression. For expression a+b, a and b are the children
	   of +.
	*/
	Children() Expressions

	/*
	   It is a utility function that takes in as input parameter
	   a mapper and maps the involved expressions to an expression.
	   If there is an error during the mapping, an error is
	   returned.
	*/
	MapChildren(mapper Mapper) error

	/*
	   This function returns an expression that is a deep copy.
	*/
	Copy() Expression
}

func (this Expressions) MapExpressions(mapper Mapper) (err error) {
	for i, e := range this {
		expr, err := mapper.Map(e)
		if err != nil {
			return err
		}

		this[i] = expr
	}

	return
}

// Expressions implements Stringer() API.
func (this Expressions) String() string {
	var exprText bytes.Buffer
	exprText.WriteString("[")
	for i, expr := range this {
		if i > 0 {
			exprText.WriteString(", ")
		}
		exprText.WriteString(expr.String())
	}
	exprText.WriteString("]")
	return exprText.String()
}

func (this Expressions) Copy() Expressions {
	rv := make(Expressions, len(this))
	copy(rv, this)
	return rv
}

func (this Expressions) EquivalentTo(other Expressions) bool {
	if len(this) != len(other) {
		return false
	}

	for i, expr := range this {
		if !expr.EquivalentTo(other[i]) {
			return false
		}
	}

	return true
}

func Copy(expr Expression) Expression {
	if expr == nil {
		return nil
	}

	return expr.Copy()
}

func CopyExpressions(exprs Expressions) Expressions {
	if exprs == nil {
		return nil
	}

	return exprs.Copy()
}

func Equivalent(expr1, expr2 Expression) bool {
	return (expr1 == nil && expr2 == nil) ||
		(expr1 != nil && expr2 != nil && expr1.EquivalentTo(expr2))
}
