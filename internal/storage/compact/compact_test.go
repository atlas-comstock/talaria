// Copyright 2019-2020 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package compact

import (
	"context"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kelindar/talaria/internal/config"

	"github.com/kelindar/talaria/internal/encoding/block"
	"github.com/kelindar/talaria/internal/encoding/key"
	"github.com/kelindar/talaria/internal/encoding/typeof"
	"github.com/kelindar/talaria/internal/monitor"
	"github.com/kelindar/talaria/internal/storage/disk"
	"github.com/stretchr/testify/assert"
)

var input = []byte{
	0xe6, 0x2, 0x2, 0x68, 0x69, 0x8, 0x0, 0x4, 0x6c, 0x69, 0x73, 0x74, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x7, 0x3, 0x6d, 0x61, 0x70, 0x9, 0x0,
	0x0, 0x0, 0x3a, 0x0, 0x0, 0x0, 0x13, 0x7, 0x8, 0x62, 0x6f, 0x6f, 0x6c, 0x65, 0x61, 0x6e, 0x31, 0x9, 0x0, 0x0, 0x0, 0x4d, 0x0, 0x0, 0x0, 0x6, 0x5, 0x5,
	0x6c, 0x6f, 0x6e, 0x67, 0x31, 0x9, 0x0, 0x0, 0x0, 0x53, 0x0, 0x0, 0x0, 0x14, 0x2, 0x6, 0x6d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x9, 0x0, 0x0, 0x0, 0x67,
	0x0, 0x0, 0x0, 0x42, 0x7, 0x7, 0x64, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x31, 0x9, 0x0, 0x0, 0x0, 0xa9, 0x0, 0x0, 0x0, 0x14, 0x3, 0x4, 0x69, 0x6e, 0x74,
	0x31, 0x9, 0x0, 0x0, 0x0, 0xbd, 0x0, 0x0, 0x0, 0x10, 0x1, 0x7, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x31, 0x9, 0x0, 0x0, 0x0, 0xcd, 0x0, 0x0, 0x0, 0x13,
	0x4, 0xe0, 0x1, 0x47, 0xc, 0x1, 0x0, 0x4, 0x0, 0x9, 0x1, 0x88, 0x38, 0x0, 0x0, 0x0, 0x38, 0x5b, 0x7b, 0x22, 0x69, 0x6e, 0x74, 0x31, 0x22, 0x3a, 0x33,
	0x2c, 0x22, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x31, 0x22, 0x3a, 0x22, 0x67, 0x6f, 0x6f, 0x64, 0x22, 0x7d, 0x2c, 0x7b, 0xd, 0x1c, 0x0, 0x34, 0x2e,
	0x1c, 0x0, 0x14, 0x62, 0x61, 0x64, 0x22, 0x7d, 0x5d, 0x11, 0x40, 0x1, 0x0, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x2, 0x7b, 0x7d,
	0x4, 0xc, 0x1, 0x0, 0x1, 0x0, 0x12, 0x44, 0x1, 0x0, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f, 0x50, 0xc,
	0x1, 0x0, 0x4, 0x0, 0x9, 0x1, 0xa0, 0x41, 0x0, 0x0, 0x0, 0x41, 0x7b, 0x22, 0x6c, 0x69, 0x73, 0x74, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x69, 0x6e, 0x74, 0x31,
	0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x31, 0x22, 0x3a, 0x22, 0x62, 0x79, 0x65, 0x22, 0x7d, 0x2c, 0x11, 0x1b, 0x0, 0x32,
	0x2e, 0x1b, 0x0, 0x1c, 0x73, 0x69, 0x67, 0x68, 0x22, 0x7d, 0x5d, 0x7d, 0x12, 0x44, 0x1, 0x0, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x2e, 0xc0, 0xe, 0x34, 0x1, 0x0, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x11, 0x40, 0x1, 0x0, 0x4, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x2, 0x68, 0x69, 0x0}

// blockWriter mock
type blockWriter func([]block.Block, typeof.Schema) error

func (w blockWriter) WriteBlock(blocks []block.Block, schema typeof.Schema) error {
	return w(blocks, schema)
}

// Opens a new disk storage and runs a a test on it.
func runTest(t *testing.T, test func(store *disk.Storage)) {
	assert.NotPanics(t, func() {
		run(test)
	})
}

// Run runs a function on a temp store
func run(f func(store *disk.Storage)) {
	dir, _ := ioutil.TempDir("", "test")
	store := disk.New(monitor.NewNoop())
	syncWrite := false
	_ = store.Open(dir, config.Badger{
		SyncWrites: &syncWrite,
	})

	// Close once we're done and delete data
	defer func() { _ = os.RemoveAll(dir) }()
	defer func() { _ = store.Close() }()

	f(store)
}

func TestRange(t *testing.T) {
	runTest(t, func(buffer *disk.Storage) {
		var count int64
		var dest blockWriter = func(blocks []block.Block, schema typeof.Schema) error {
			atomic.AddInt64(&count, 1)
			return nil
		}

		store := New(buffer, dest, monitor.NewNoop(), 100*time.Millisecond)

		// Insert out of order
		_ = store.Append(key.New("A", time.Unix(0, 0)), input, 60*time.Second)
		_ = store.Append(key.New("A", time.Unix(1, 0)), input, 60*time.Second)
		_ = store.Append(key.New("C", time.Unix(1, 0)), input, 60*time.Second)
		_ = store.Append(key.New("B", time.Unix(0, 0)), input, 60*time.Second)
		_ = store.Append(key.New("B", time.Unix(1, 0)), input, 60*time.Second)
		_ = store.Append(key.New("B", time.Unix(2, 0)), input, 60*time.Second)
		_ = store.Append(key.New("D", time.Unix(2, 0)), input, 60*time.Second)

		// Iterate in order
		var values [][]byte
		err := store.Range(key.First(), key.Last(), func(k, v []byte) bool {
			values = append(values, v)
			return false
		})

		// Must be in order
		assert.NoError(t, err)
		assert.Equal(t, 7, len(values))
		for _, v := range values {
			assert.EqualValues(t, input, v)
		}

		// Manually compact, the final count should be 2 (given we have 2 keys)
		store.Compact(context.Background())
		assert.Equal(t, int64(4), count)
	})
}
