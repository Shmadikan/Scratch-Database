package diskdata

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shmadikan/Scratch-Database/internal/btree"
)


func generateStore(rootNumber int, flushedPages int) Store {
	return Store{
		Path: "/",
		fileDescriptor: 0,
		tree: btree.Btree{RootPageNumber: uint64(rootNumber)},
		page: struct{flushed uint64; temp [][]byte}{
			flushed: uint64(flushedPages),
		},
	}
} 


func TestMetaData(t *testing.T) {
	tests := []struct{
		numberPage int
		flushedCount int
		testName string
	}{
		{
			numberPage: 2,
			flushedCount: 1,
			testName: "test1",
		},
		{
			numberPage: 2999,
			flushedCount: 432,
			testName: "test2",

		},
		{
			numberPage: 432,
			flushedCount: 199999,
			testName: "test3",

		},
		{
			numberPage: 2,
			flushedCount: 13432,
			testName: "test4",

		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
		s := generateStore(test.numberPage, test.flushedCount)
		byteMeta := saveMeta(&s)
		assert.Equal(t, test.numberPage, int(binary.LittleEndian.Uint64(byteMeta[16:])))
		assert.Equal(t, test.flushedCount, int(binary.LittleEndian.Uint64(byteMeta[24:])), "Not equal page counts")
		assert.Equal(t, DB_SIG, string(byteMeta[:16]))
		})
	}
}


func TestLoadMeta(t *testing.T) {
	s := generateStore(1, 1)
	dataStore := saveMeta(&s)
	copyData := make([]byte, 32)
	copy(copyData, dataStore)
	assert.Exactly(t, dataStore, copyData)
}