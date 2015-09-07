//  Copyright (c) 2015 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

// +build ignore

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const fuzzPrefix = "workdir/corpus"
const tutorialDir = "../../tutorial/content/"

var startBytes = []byte(`<pre id="example">`)
var endBytes = []byte(`</pre>`)

func main() {

	os.MkdirAll(fuzzPrefix, 0777)
	fileInfos, err := ioutil.ReadDir(tutorialDir)
	if err != nil {
		log.Fatal(err)
	}
	i := 1
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".md") {
			fileBytes, err := ioutil.ReadFile(tutorialDir + string(os.PathSeparator) + fileInfo.Name())
			if err != nil {
				log.Fatal(err)
			}
			startPos := bytes.Index(fileBytes, startBytes)
			if startPos < 0 {
				log.Fatalf("cannot find start position for %s", fileInfo.Name())
				continue
			}
			fileBytes = fileBytes[startPos+len(startBytes):]
			endPos := bytes.Index(fileBytes, endBytes)
			if endPos < 0 {
				log.Fatalf("cannot find end position for %s", fileInfo.Name())
			}
			fileBytes = fileBytes[:endPos]
			fileBytes = bytes.TrimSpace(fileBytes)
			ioutil.WriteFile(fuzzPrefix+string(os.PathSeparator)+strconv.Itoa(i)+".txt", fileBytes, 0777)
			i++
		}
	}
}
