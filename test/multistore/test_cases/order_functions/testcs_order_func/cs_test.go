//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package testcs_order_func

import (
	"fmt"
	"path/filepath"
	"testing"
)

/*
Insert data into the orders bucket created earlier
using the statements in insert_orders.json.
*/
func TestInsertCaseFiles(t *testing.T) {
	fmt.Println("\n\nInserting values into Bucket for Order Functions \n\n ")
	qc := Start_test()
	matches, err := filepath.Glob("../insert.json")
	if err != nil {
		t.Errorf("glob failed: %v", err)
	}
	for _, m := range matches {
		t.Logf("TestCaseFile: %v\n", m)
		stmt, err := testCaseFile(m, qc)
		if err != nil {
			t.Errorf("Error received : %s \n", err)
			return
		}
		if stmt != "" {
			t.Logf(" %v\n", stmt)
		}
		fmt.Println("\nQuery matched: ", m, "\n\n")
	}
}

func TestAllCaseFiles(t *testing.T) {
	qc := Start_test()
	matches, err := filepath.Glob("../case_*.json")
	if err != nil {
		t.Errorf("glob failed: %v", err)
	}
	for _, m := range matches {
		t.Logf("TestCaseFile: %v\n", m)
		stmt, err := testCaseFile(m, qc)
		if err != nil {
			t.Errorf("Error received : %s \n", err)
			return
		}
		if stmt != "" {
			t.Logf(" %v\n", stmt)
		}
		fmt.Println("\nQuery matched: ", m, "\n\n")
	}
}

func TestCleanupData(t *testing.T) {
	qc := Start_test()

	_, _, errcs := Run_test(qc, "delete from orders where test_id = \"order_func\"")
	if errcs != nil {
		t.Errorf("did not expect err %s", errcs.Error())
	}

}
