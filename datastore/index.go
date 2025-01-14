//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package datastore

import (
	atomic "github.com/couchbase/go-couchbase/platform"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/timestamp"
	"github.com/couchbase/query/value"
)

type IndexType string

const (
	DEFAULT IndexType = "default" // default may vary per backend
	VIEW    IndexType = "view"    // view index
	GSI     IndexType = "gsi"     // global secondary index
)

type Indexer interface {
	KeyspaceId() string                                                                       // Id of the keyspace to which this indexer belongs
	Name() IndexType                                                                          // Unique within a Keyspace.
	IndexIds() ([]string, errors.Error)                                                       // Ids of the indexes defined on this keyspace
	IndexNames() ([]string, errors.Error)                                                     // Names of the indexes defined on this keyspace
	IndexById(id string) (Index, errors.Error)                                                // Find an index on this keyspace using the index's id
	IndexByName(name string) (Index, errors.Error)                                            // Find an index on this keyspace using the index's name
	PrimaryIndexes() ([]PrimaryIndex, errors.Error)                                           // Returns the server-recommended primary index
	Indexes() ([]Index, errors.Error)                                                         // Returns all the indexes defined on this keyspace
	CreatePrimaryIndex(requestId, name string, with value.Value) (PrimaryIndex, errors.Error) // Create or return a primary index on this keyspace
	CreateIndex(requestId, name string, equalKey, rangeKey expression.Expressions,            // Create a secondary index on this keyspace
		where expression.Expression, with value.Value) (Index, errors.Error)
	BuildIndexes(requestId string, name ...string) errors.Error // Build indexes that were deferred at creation
	Refresh() errors.Error                                      // Refresh list of indexes from metadata
	SetLogLevel(level logging.Level)                            // Set log level for in-process logging
}

type IndexState string

const (
	DEFERRED IndexState = "deferred" // The index has not been built
	BUILDING IndexState = "building" // The index is being built or rebuilt
	PENDING  IndexState = "pending"  // The index is in progress but is not yet ready for use
	ONLINE   IndexState = "online"   // The index is available for use
	OFFLINE  IndexState = "offline"  // The index requires manual intervention
)

func (indexState IndexState) String() string {
	return string(indexState)
}

type ScanConsistency string

const (
	UNBOUNDED ScanConsistency = "unbounded"
	SCAN_PLUS ScanConsistency = "scan_plus"
	AT_PLUS   ScanConsistency = "at_plus"
)

type IndexKey expression.Expressions

type Indexes []Index

/*
Index is the base type for indexes, which may be distributed.
*/
type Index interface {
	KeyspaceId() string                                                 // Id of the keyspace to which this index belongs
	Id() string                                                         // Id of this index
	Name() string                                                       // Name of this index
	Type() IndexType                                                    // Type of this index
	SeekKey() expression.Expressions                                    // Equality keys
	RangeKey() expression.Expressions                                   // Range keys
	Condition() expression.Expression                                   // Condition, if any
	IsPrimary() bool                                                    // Is this a primary index
	State() (state IndexState, msg string, err errors.Error)            // Obtain state of this index
	Statistics(requestId string, span *Span) (Statistics, errors.Error) // Obtain statistics for this index
	Drop(requestId string) errors.Error                                 // Drop / delete this index
	Scan(requestId string, span *Span, distinct bool, limit int64, cons ScanConsistency,
		vector timestamp.Vector, conn *IndexConnection) // Perform a scan on this index. Distinct and limit are hints.
}

/*
PrimaryIndex represents primary key indexes.
*/
type PrimaryIndex interface {
	Index
	ScanEntries(requestId string, limit int64, cons ScanConsistency,
		vector timestamp.Vector, conn *IndexConnection) // Perform a scan of all the entries in this index
}

type SizedIndex interface {
	Index
	SizeFromStatistics(requestId string) (int64, errors.Error)
}

type Range struct {
	Low       value.Values
	High      value.Values
	Inclusion Inclusion
}

type Ranges []*Range

// Inclusion controls how the boundary values of a range are treated.
type Inclusion int

const (
	NEITHER Inclusion = 0x00
	LOW               = 0x01
	HIGH              = 0x01 << 1
	BOTH              = LOW | HIGH
)

type Span struct {
	Seek  value.Values
	Range Range
}

type Spans []*Span

type IndexEntry struct {
	EntryKey   value.Values
	PrimaryKey string
}

type EntryChannel chan *IndexEntry
type StopChannel chan bool

// Statistics captures statistics for a range.
// - it may return heuristics and/or outdated values.
// - query shall not depend on the accuracy of this statistics.
// - primarily intended for optimizer's consumption.
type Statistics interface {
	Count() (int64, errors.Error)
	Min() (value.Values, errors.Error)
	Max() (value.Values, errors.Error)
	DistinctCount() (int64, errors.Error)
	Bins() ([]Statistics, errors.Error)
}

type Context interface {
	Fatal(errors.Error)
	Error(errors.Error)
	Warning(errors.Error)
}

type IndexConnection struct {
	entryChannel EntryChannel // Closed by the index when the scan is completed or aborted.
	stopChannel  StopChannel  // Notifies index to stop scanning. Never closed, just garbage-collected.
	context      Context
	timeout      bool
	primary      bool
}

const _ENTRY_CAP = 256 // Index scan request size

func NewIndexConnection(context Context) *IndexConnection {
	return &IndexConnection{
		entryChannel: make(EntryChannel, _ENTRY_CAP),
		stopChannel:  make(StopChannel, 1),
		context:      context,
	}
}

var scanCap atomic.AlignedInt64

func SetScanCap(cap int64) {
	atomic.StoreInt64(&scanCap, cap)
}

func GetScanCap() int64 {
	return atomic.LoadInt64(&scanCap)
}

func NewSizedIndexConnection(size int64, context Context) (*IndexConnection, errors.Error) {
	if size <= 0 {
		return nil, errors.NewIndexScanSizeError(size)
	}
	maxSize := GetScanCap()
	if (maxSize > 0) && (size > maxSize) {
		size = maxSize
	}

	return &IndexConnection{
		entryChannel: make(EntryChannel, size),
		stopChannel:  make(StopChannel, 1),
		context:      context,
	}, nil
}

func (this *IndexConnection) EntryChannel() EntryChannel {
	return this.entryChannel
}

func (this *IndexConnection) StopChannel() StopChannel {
	return this.stopChannel
}

func (this *IndexConnection) Fatal(err errors.Error) {
	this.context.Fatal(err)
}

func (this *IndexConnection) Error(err errors.Error) {
	if this.primary && err.Code() == errors.INDEX_SCAN_TIMEOUT {
		this.timeout = true
		return
	}
	this.context.Error(err)
}

func (this *IndexConnection) Warning(wrn errors.Error) {
	this.context.Warning(wrn)
}

func (this *IndexConnection) SetPrimary() {
	this.primary = true
}

func (this *IndexConnection) Timeout() bool {
	return this.timeout
}
