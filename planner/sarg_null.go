//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package planner

import (
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
)

var _NULL_SPANS plan.Spans

func init() {
	span := &plan.Span{}
	span.Range.Low = expression.Expressions{expression.NULL_EXPR}
	span.Range.High = span.Range.Low
	span.Range.Inclusion = datastore.BOTH
	_NULL_SPANS = plan.Spans{span}
}

type sargNull struct {
	sargBase
}

func newSargNull(pred *expression.IsNull) *sargNull {
	rv := &sargNull{}
	rv.sarger = func(expr2 expression.Expression) (plan.Spans, error) {
		if SubsetOf(pred, expr2) {
			return _SELF_SPANS, nil
		}

		if pred.Operand().EquivalentTo(expr2) {
			return _NULL_SPANS, nil
		}

		return nil, nil
	}

	return rv
}

type sargNotNull struct {
	sargBase
}

func newSargNotNull(pred *expression.IsNotNull) *sargNotNull {
	rv := &sargNotNull{}
	rv.sarger = func(expr2 expression.Expression) (plan.Spans, error) {
		if SubsetOf(pred, expr2) {
			return _SELF_SPANS, nil
		}

		if pred.Operand().EquivalentTo(expr2) {
			return _VALUED_SPANS, nil
		}

		return nil, nil
	}

	return rv
}
