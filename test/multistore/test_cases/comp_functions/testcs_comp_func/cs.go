//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.
package testcs_comp_func

import (
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/server"
	js "github.com/couchbase/query/test/multistore"
)

func Start_test() *server.Server {
	return js.Start(js.Site_CBS, js.Auth_param+"@"+js.Pool_CBS, js.Namespace_CBS)
}

func testCaseFile(fname string, qc *server.Server) (fin_stmt string, errstring error) {
	fin_stmt, errstring = js.FtestCaseFile(fname, qc, js.Namespace_CBS)
	return
}

func Run_test(mockServer *server.Server, q string) ([]interface{}, []errors.Error, errors.Error) {
	return js.Run(mockServer, q, js.Namespace_CBS)
}
