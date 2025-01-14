//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

/*

Package mock provides a fake, mock 100%-in-memory implementation of
the datastore package, which can be useful for testing.  Because it is
memory-oriented, performance testing of higher layers may be easier
with this mock datastore.

*/
package mock

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/timestamp"
	"github.com/couchbase/query/value"
)

const (
	DEFAULT_NUM_NAMESPACES = 1
	DEFAULT_NUM_KEYSPACES  = 1
	DEFAULT_NUM_ITEMS      = 100000
)

// store is the root for the mock-based Store.
type store struct {
	path           string
	namespaces     map[string]*namespace
	namespaceNames []string
	params         map[string]int
}

func (s *store) Id() string {
	return s.URL()
}

func (s *store) URL() string {
	return "mock:" + s.path
}

func (s *store) NamespaceIds() ([]string, errors.Error) {
	return s.NamespaceNames()
}

func (s *store) NamespaceNames() ([]string, errors.Error) {
	return s.namespaceNames, nil
}

func (s *store) NamespaceById(id string) (p datastore.Namespace, e errors.Error) {
	return s.NamespaceByName(id)
}

func (s *store) NamespaceByName(name string) (p datastore.Namespace, e errors.Error) {
	p, ok := s.namespaces[name]
	if !ok {
		p, e = nil, errors.NewOtherNamespaceNotFoundError(nil, name+" for Mock datastore")
	}

	return
}

func (s *store) Authorize(datastore.Privileges, datastore.Credentials) errors.Error {
	return nil
}

func (s *store) SetLogLevel(level logging.Level) {
	// No-op. Uses query engine logger.
}

// namespace represents a mock-based Namespace.
type namespace struct {
	store         *store
	name          string
	keyspaces     map[string]*keyspace
	keyspaceNames []string
}

func (p *namespace) DatastoreId() string {
	return p.store.Id()
}

func (p *namespace) Id() string {
	return p.Name()
}

func (p *namespace) Name() string {
	return p.name
}

func (p *namespace) KeyspaceIds() ([]string, errors.Error) {
	return p.KeyspaceNames()
}

func (p *namespace) KeyspaceNames() ([]string, errors.Error) {
	return p.keyspaceNames, nil
}

func (p *namespace) KeyspaceById(id string) (b datastore.Keyspace, e errors.Error) {
	return p.KeyspaceByName(id)
}

func (p *namespace) KeyspaceByName(name string) (b datastore.Keyspace, e errors.Error) {
	b, ok := p.keyspaces[name]
	if !ok {
		b, e = nil, errors.NewOtherKeyspaceNotFoundError(nil, name+" for Mock datastore")
	}

	return
}

// keyspace is a mock-based keyspace.
type keyspace struct {
	namespace *namespace
	name      string
	nitems    int
	mi        datastore.Indexer
}

func (b *keyspace) NamespaceId() string {
	return b.namespace.Id()
}

func (b *keyspace) Id() string {
	return b.Name()
}

func (b *keyspace) Name() string {
	return b.name
}

func (b *keyspace) Count() (int64, errors.Error) {
	return int64(b.nitems), nil
}

func (b *keyspace) Indexer(name datastore.IndexType) (datastore.Indexer, errors.Error) {
	return b.mi, nil
}

func (b *keyspace) Indexers() ([]datastore.Indexer, errors.Error) {
	return []datastore.Indexer{b.mi}, nil
}

func (b *keyspace) Fetch(keys []string) ([]datastore.AnnotatedPair, []errors.Error) {
	var errs []errors.Error
	rv := make([]datastore.AnnotatedPair, 0, len(keys))
	for _, k := range keys {
		item, e := b.fetchOne(k)
		if e != nil {
			if errs == nil {
				errs = make([]errors.Error, 0, 1)
			}
			errs = append(errs, e)
			continue
		}

		if item != nil {
			item.SetAttachment("meta", map[string]interface{}{
				"id": k,
			})
		}

		rv = append(rv, datastore.AnnotatedPair{
			Key:   k,
			Value: item,
		})
	}
	return rv, errs
}

func (b *keyspace) fetchOne(key string) (value.AnnotatedValue, errors.Error) {
	i, e := strconv.Atoi(key)
	if e != nil {
		return nil, errors.NewOtherKeyNotFoundError(e, fmt.Sprintf("no mock item: %v", key))
	} else {
		return genItem(i, b.nitems)
	}
}

// generate a mock document - used by fetchOne to mock a document in the keyspace
func genItem(i int, nitems int) (value.AnnotatedValue, errors.Error) {
	if i < 0 || i >= nitems {
		return nil, errors.NewOtherDatastoreError(nil,
			fmt.Sprintf("item out of mock range: %v [0,%v)", i, nitems))
	}
	id := strconv.Itoa(i)
	doc := value.NewAnnotatedValue(map[string]interface{}{"id": id, "i": float64(i)})
	doc.SetAttachment("meta", map[string]interface{}{"id": id})
	return doc, nil
}

func (b *keyspace) Insert(inserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewOtherNotImplementedError(nil, "for Mock datastore")
}

func (b *keyspace) Update(updates []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewOtherNotImplementedError(nil, "for Mock datastore")
}

func (b *keyspace) Upsert(upserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewOtherNotImplementedError(nil, "for Mock datastore")
}

func (b *keyspace) Delete(deletes []string) ([]string, errors.Error) {
	// FIXME
	return nil, errors.NewOtherNotImplementedError(nil, "for Mock datastore")
}

func (b *keyspace) Release() {
}

type mockIndexer struct {
	keyspace *keyspace
	indexes  map[string]datastore.Index
	primary  datastore.PrimaryIndex
}

func newMockIndexer(keyspace *keyspace) datastore.Indexer {

	return &mockIndexer{
		keyspace: keyspace,
		indexes:  make(map[string]datastore.Index),
	}
}

func (mi *mockIndexer) KeyspaceId() string {
	return mi.keyspace.Id()
}

func (mi *mockIndexer) Name() datastore.IndexType {
	return datastore.DEFAULT
}

func (mi *mockIndexer) IndexIds() ([]string, errors.Error) {
	rv := make([]string, 0, len(mi.indexes))
	for name, _ := range mi.indexes {
		rv = append(rv, name)
	}
	return rv, nil
}

func (mi *mockIndexer) IndexNames() ([]string, errors.Error) {
	rv := make([]string, 0, len(mi.indexes))
	for name, _ := range mi.indexes {
		rv = append(rv, name)
	}
	return rv, nil
}

func (mi *mockIndexer) IndexById(id string) (datastore.Index, errors.Error) {
	return mi.IndexByName(id)
}

func (mi *mockIndexer) IndexByName(name string) (datastore.Index, errors.Error) {
	index, ok := mi.indexes[name]
	if !ok {
		return nil, errors.NewOtherIdxNotFoundError(nil, name+"for Mock datastore")
	}
	return index, nil
}

func (mi *mockIndexer) PrimaryIndexes() ([]datastore.PrimaryIndex, errors.Error) {
	return []datastore.PrimaryIndex{mi.primary}, nil
}

func (mi *mockIndexer) Indexes() ([]datastore.Index, errors.Error) {
	return []datastore.Index{mi.primary}, nil
}

func (mi *mockIndexer) CreatePrimaryIndex(requestId, name string, with value.Value) (datastore.PrimaryIndex, errors.Error) {
	if mi.primary == nil {
		pi := new(primaryIndex)
		mi.primary = pi
		pi.keyspace = mi.keyspace
		pi.name = name
		mi.indexes[pi.name] = pi
	}

	return mi.primary, nil
}

func (mi *mockIndexer) CreateIndex(requestId, name string, equalKey, rangeKey expression.Expressions,
	where expression.Expression, with value.Value) (datastore.Index, errors.Error) {
	return nil, errors.NewOtherNotSupportedError(nil, "CREATE INDEX is not supported for mock datastore.")
}

func (mi *mockIndexer) BuildIndexes(requestId string, names ...string) errors.Error {
	return errors.NewOtherNotSupportedError(nil, "BUILD INDEXES is not supported for mock datastore.")
}

func (mi *mockIndexer) Refresh() errors.Error {
	return nil
}

func (mi *mockIndexer) SetLogLevel(level logging.Level) {
	// No-op, uses query engine logger
}

// NewDatastore creates a new mock store for the given "path".  The
// path has prefix "mock:", with the rest of the path treated as a
// comma-separated key=value params.  For example:
// mock:namespaces=2,keyspaces=5,items=50000 The above means 2
// namespaces.  And, each namespace has 5 keyspaces.  And, each
// keyspace with 50000 items.  By default, you get...
// mock:namespaces=1,keyspaces=1,items=100000 Which is what you'd get
// by specifying a path of just...  mock:
func NewDatastore(path string) (datastore.Datastore, errors.Error) {
	if strings.HasPrefix(path, "mock:") {
		path = path[5:]
	}
	params := map[string]int{}
	for _, kv := range strings.Split(path, ",") {
		if kv == "" {
			continue
		}
		pair := strings.Split(kv, "=")
		v, e := strconv.Atoi(pair[1])
		if e != nil {
			return nil, errors.NewOtherDatastoreError(e,
				fmt.Sprintf("could not parse mock param key: %s, val: %s",
					pair[0], pair[1]))
		}
		params[pair[0]] = v
	}
	nnamespaces := paramVal(params, "namespaces", DEFAULT_NUM_NAMESPACES)
	nkeyspaces := paramVal(params, "keyspaces", DEFAULT_NUM_KEYSPACES)
	nitems := paramVal(params, "items", DEFAULT_NUM_ITEMS)
	s := &store{path: path, params: params, namespaces: map[string]*namespace{}, namespaceNames: []string{}}
	for i := 0; i < nnamespaces; i++ {
		p := &namespace{store: s, name: "p" + strconv.Itoa(i), keyspaces: map[string]*keyspace{}, keyspaceNames: []string{}}
		for j := 0; j < nkeyspaces; j++ {
			b := &keyspace{namespace: p, name: "b" + strconv.Itoa(j), nitems: nitems}

			b.mi = newMockIndexer(b)
			b.mi.CreatePrimaryIndex("", "#primary", nil)
			p.keyspaces[b.name] = b
			p.keyspaceNames = append(p.keyspaceNames, b.name)
		}
		s.namespaces[p.name] = p
		s.namespaceNames = append(s.namespaceNames, p.name)
	}
	return s, nil
}

func paramVal(params map[string]int, key string, defaultVal int) int {
	v, ok := params[key]
	if ok {
		return v
	}
	return defaultVal
}

// primaryIndex performs full keyspace scans.
type primaryIndex struct {
	name     string
	keyspace *keyspace
}

func (pi *primaryIndex) KeyspaceId() string {
	return pi.keyspace.Id()
}

func (pi *primaryIndex) Id() string {
	return pi.Name()
}

func (pi *primaryIndex) Name() string {
	return pi.name
}

func (pi *primaryIndex) Type() datastore.IndexType {
	return datastore.DEFAULT
}

func (pi *primaryIndex) SeekKey() expression.Expressions {
	return nil
}

func (pi *primaryIndex) RangeKey() expression.Expressions {
	return nil
}

func (pi *primaryIndex) Condition() expression.Expression {
	return nil
}

func (pi *primaryIndex) IsPrimary() bool {
	return true
}

func (pi *primaryIndex) State() (state datastore.IndexState, msg string, err errors.Error) {
	return datastore.ONLINE, "", nil
}

func (pi *primaryIndex) Statistics(requestId string, span *datastore.Span) (
	datastore.Statistics, errors.Error) {
	return nil, nil
}

func (pi *primaryIndex) Drop(requestId string) errors.Error {
	return errors.NewOtherIdxNoDrop(nil, "This primary index cannot be dropped for Mock datastore.")
}

func (pi *primaryIndex) Scan(requestId string, span *datastore.Span, distinct bool, limit int64,
	cons datastore.ScanConsistency, vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())

	// For primary indexes, bounds must always be strings, so we
	// can just enforce that directly
	low, high := "", ""

	// Ensure that lower bound is a string, if any
	if len(span.Range.Low) > 0 {
		a := span.Range.Low[0].Actual()
		switch a := a.(type) {
		case string:
			low = a
		default:
			conn.Error(errors.NewOtherDatastoreError(nil, fmt.Sprintf("Invalid lower bound %v of type %T.", a, a)))
			return
		}
	}

	// Ensure that upper bound is a string, if any
	if len(span.Range.High) > 0 {
		a := span.Range.High[0].Actual()
		switch a := a.(type) {
		case string:
			high = a
		default:
			conn.Error(errors.NewOtherDatastoreError(nil, fmt.Sprintf("Invalid upper bound %v of type %T.", a, a)))
			return
		}
	}

	if limit == 0 {
		limit = int64(pi.keyspace.nitems)
	}

	for i := 0; i < pi.keyspace.nitems && int64(i) < limit; i++ {
		id := strconv.Itoa(i)

		if low != "" &&
			(id < low ||
				(id == low && (span.Range.Inclusion&datastore.LOW == 0))) {
			continue
		}

		low = ""

		if high != "" &&
			(id > high ||
				(id == high && (span.Range.Inclusion&datastore.HIGH == 0))) {
			break
		}

		entry := datastore.IndexEntry{PrimaryKey: id}
		conn.EntryChannel() <- &entry
	}
}

func (pi *primaryIndex) ScanEntries(requestId string, limit int64, cons datastore.ScanConsistency,
	vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())

	if limit == 0 {
		limit = int64(pi.keyspace.nitems)
	}

	for i := 0; i < pi.keyspace.nitems && int64(i) < limit; i++ {
		entry := datastore.IndexEntry{PrimaryKey: strconv.Itoa(i)}
		conn.EntryChannel() <- &entry
	}
}
