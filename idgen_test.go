package idgen

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

var m map[uint64]uint64 = make(map[uint64]uint64)

func TestIdGen(t *testing.T) {
	id := NewIdgen(2, 1, 0, 0)
	go id.Recv()
	for i := 0; i < 10000; i++ {
		go newid(t, id)
	}
	time.Sleep(2 * time.Second)
	assert.Equal(t, 1, 1)
	id.Stop()

}

func newid(t *testing.T, id *IdGen) {
	rand.Seed(time.Now().UnixNano())

	NewId := id.GetNewId()
	_, ok := m[NewId]
	assert.False(t, ok) // 如果map 中已经有此newid ,则说明， 生成的主键冲突
	m[NewId] = NewId
}
