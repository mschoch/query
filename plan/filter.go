//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package plan

import (
	"encoding/json"

	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/expression/parser"
)

type Filter struct {
	readonly
	cond expression.Expression
}

func NewFilter(cond expression.Expression) *Filter {
	return &Filter{
		cond: cond,
	}
}

func (this *Filter) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFilter(this)
}

func (this *Filter) New() Operator {
	return &Filter{}
}

func (this *Filter) Condition() expression.Expression {
	return this.cond
}

func (this *Filter) MarshalJSON() ([]byte, error) {
	r := map[string]interface{}{"#operator": "Filter"}
	r["condition"] = expression.NewStringer().Visit(this.cond)
	return json.Marshal(r)
}

func (this *Filter) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_         string `json:"#operator"`
		Condition string `json:"condition"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	if _unmarshalled.Condition != "" {
		this.cond, err = parser.Parse(_unmarshalled.Condition)
	}

	return err
}
