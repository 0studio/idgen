package idgen

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

var m3 map[uint64]uint64 = make(map[uint64]uint64)

func TestIdGen31(t *testing.T) {
	var (
		platformBits uint64 = 3
		serverBits   uint64 = 3
		sysTypeBits  uint64 = 3
		platform     uint64 = 1
		server       uint64 = 1
		sysType      uint64 = 1
	)
	idGen := NewIdgen3(platformBits, platform,
		serverBits, server,
		sysTypeBits, sysType)
	assert.True(t, idGen.GetSeqBits() > 3)

}
func TestIdGen32(t *testing.T) {
	var (
		// platformBits,serverBits,sysTypeBits 所占数据 可以为0
		platformBits uint64 = 0
		serverBits   uint64 = 0
		sysTypeBits  uint64 = 10
		platform     uint64 = 0
		server       uint64 = 0
		sysType      uint64 = 1
	)
	idGen := NewIdgen3(platformBits, platform,
		serverBits, server,
		sysTypeBits, sysType)
	assert.True(t, idGen.GetSeqBits() > 3)

}

//
func TestIdGen3(t *testing.T) {
	id := NewIdgen3(8, 2, 4, 1, 8, 0)
	for j := 0; j < 1000; j++ {
		go func() {
			for i := 0; i < 1000; i++ {
				go newid3(t, id, i, j)
			}
		}()
	}

	time.Sleep(10 * time.Second)
	assert.Equal(t, 1, 1)

}

func newid3(t *testing.T, id *IdGen3, i, j int) {
	rand.Seed(time.Now().UnixNano())

	NewId := id.GetNewId()
	old, ok := m3[NewId]
	// fmt.Println(NewId, i, j)
	assert.False(t, ok, strconv.FormatUint(old, 10)) // 如果map 中已经有此newid ,则说明， 生成的主键冲突
	m3[NewId] = NewId
}
