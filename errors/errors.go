//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

/*

Package err provides user-visible errors and warnings. These errors
include error codes and will eventually provide multi-language
messages.

*/
package errors

import (
	"encoding/json"
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	EXCEPTION = iota
	WARNING
	NOTICE
	INFO
	LOG
	DEBUG
)

type Errors []Error

// Error will eventually include code, message key, and internal error
// object (cause) and message
type Error interface {
	error
	Code() int32
	TranslationKey() string
	Cause() error
	Level() int
	IsFatal() bool
}

type ErrorChannel chan Error

func NewError(e error, internalMsg string) Error {
	switch e := e.(type) {
	case Error: // if given error is already an Error, just return it:
		return e
	default:
		return &err{level: EXCEPTION, ICode: 5000, IKey: "Internal Error", ICause: e,
			InternalMsg: internalMsg, InternalCaller: CallerN(1)}
	}
}

func NewWarning(internalMsg string) Error {
	return &err{level: WARNING, InternalMsg: internalMsg, InternalCaller: CallerN(1)}
}

func NewNotice(internalMsg string) Error {
	return &err{level: NOTICE, InternalMsg: internalMsg, InternalCaller: CallerN(1)}
}

func NewInfo(internalMsg string) Error {
	return &err{level: INFO, InternalMsg: internalMsg, InternalCaller: CallerN(1)}
}

func NewLog(internalMsg string) Error {
	return &err{level: LOG, InternalMsg: internalMsg, InternalCaller: CallerN(1)}
}

func NewDebug(internalMsg string) Error {
	return &err{level: DEBUG, InternalMsg: internalMsg, InternalCaller: CallerN(1)}
}

type err struct {
	ICode          int32
	IKey           string
	ICause         error
	InternalMsg    string
	InternalCaller string
	level          int
}

func (e *err) Error() string {
	switch {
	default:
		return "Unspecified error."
	case e.InternalMsg != "" && e.ICause != nil:
		return e.InternalMsg + " - cause: " + e.ICause.Error()
	case e.InternalMsg != "":
		return e.InternalMsg
	case e.ICause != nil:
		return e.ICause.Error()
	}
}

func (e *err) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"code":    e.ICode,
		"key":     e.IKey,
		"message": e.InternalMsg,
	}
	if e.ICause != nil {
		m["cause"] = e.ICause.Error()
	}
	if e.InternalCaller != "" &&
		!strings.HasPrefix("e.InternalCaller", "unknown:") {
		m["caller"] = e.InternalCaller
	}
	return json.Marshal(m)
}

func (e *err) Level() int {
	return e.level
}

func (e *err) IsFatal() bool {
	if e.level == EXCEPTION {
		return true
	}
	return false
}

func (e *err) Code() int32 {
	return e.ICode
}

func (e *err) TranslationKey() string {
	return e.IKey
}

func (e *err) Cause() error {
	return e.ICause
}

func NewParseError(e error, msg string) Error {
	return &err{level: EXCEPTION, ICode: 4100, IKey: "parse_error", ICause: e, InternalMsg: msg, InternalCaller: CallerN(1)}
}

func NewSemanticError(e error, msg string) Error {
	return &err{level: EXCEPTION, ICode: 4200, IKey: "semantic_error", ICause: e, InternalMsg: msg, InternalCaller: CallerN(1)}
}

func NewBucketDoesNotExist(bucket string) Error {
	return &err{level: EXCEPTION, ICode: 4040, IKey: "bucket_not_found", InternalMsg: fmt.Sprintf("Bucket %s does not exist", bucket), InternalCaller: CallerN(1)}
}

func NewPoolDoesNotExist(pool string) Error {
	return &err{level: EXCEPTION, ICode: 4041, IKey: "pool_not_found", InternalMsg: fmt.Sprintf("Pool %s does not exist", pool), InternalCaller: CallerN(1)}
}

func NewTimeoutError(timeout *time.Duration) Error {
	return &err{level: EXCEPTION, ICode: 4080, IKey: "timeout", InternalMsg: fmt.Sprintf("Timeout %v exceeded", timeout), InternalCaller: CallerN(1)}
}

func NewTotalRowsInfo(rows int) Error {
	return &err{level: INFO, ICode: 100, IKey: "total_rows", InternalMsg: fmt.Sprintf("%d", rows), InternalCaller: CallerN(1)}
}

func NewTotalElapsedTimeInfo(time string) Error {
	return &err{level: INFO, ICode: 101, IKey: "total_elapsed_time", InternalMsg: fmt.Sprintf("%s", time), InternalCaller: CallerN(1)}
}

func NewNotImplemented(feature string) Error {
	return &err{level: EXCEPTION, ICode: 1001, IKey: "not_implemented", InternalMsg: fmt.Sprintf("Not yet implemented: %v", feature), InternalCaller: CallerN(1)}
}

// Returns "FileName:LineNum" of caller.
func Caller() string {
	return CallerN(1)
}

// Returns "FileName:LineNum" of the Nth caller on the call stack,
// where level of 0 is the caller of CallerN.
func CallerN(level int) string {
	_, fname, lineno, ok := runtime.Caller(1 + level)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d",
		strings.Split(path.Base(fname), ".")[0], lineno)
}
