package idgen

import (
	"github.com/0studio/bit"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
	"time"
)

var m map[uint64]uint64 = make(map[uint64]uint64)

func TestIdGen1(t *testing.T) {
	var (
		platformBits uint64 = 3
		serverBits   uint64 = 3
		sysTypeBits  uint64 = 3
		platform     uint64 = 1
		server       uint64 = 1
		sysType      uint64 = 1
		seq          uint64 = 1
	)
	idGen := NewIdgen(platformBits, platform,
		serverBits, server,
		sysTypeBits, sysType,
		seq)
	assert.Equal(t, idGen.GetSeqBits(), 54)
	assert.Equal(t, uint64(math.Pow(2, 54)), idGen.GetMaxSequence())

}

//pos=5 ,则把数值的低5位全置成1
func GetMask(pos uint64) (value uint64) {
	var b bit.BitInt
	var i uint64
	for i = 0; i < pos; i++ {
		b.SetFlag(i)
	}
	return uint64(b)

}
func TestIdGenSequenceMask(t *testing.T) {
	var (
		platformBits uint64 = 30
		serverBits   uint64 = 30
		sysTypeBits  uint64 = 1
		platform     uint64 = 1
		server       uint64 = 1
		sysType      uint64 = 1
		seq          uint64 = 1
	)
	idGen := NewIdgen(platformBits, platform,
		serverBits, server,
		sysTypeBits, sysType,
		seq)
	assert.Equal(t, idGen.GetSeqBits(), 2)
	assert.Equal(t, 3, idGen.GetSequenceMask()) // 3=2#11

}

func TestIdGenPlatformMask(t *testing.T) {
	var (
		platformBits uint64 = 63
		serverBits   uint64 = 0
		sysTypeBits  uint64 = 0
		platform     uint64 = 1
		server       uint64 = 1
		sysType      uint64 = 1
		seq          uint64 = 1
	)
	idGen := NewIdgen(platformBits, platform,
		serverBits, server,
		sysTypeBits, sysType,
		seq)
	assert.Equal(t, GetMask(63), idGen.GetPlatformMask())
}

func TestIdGen2(t *testing.T) {
	var (
		platformBits uint64 = 12
		serverBits   uint64 = 9
		sysTypeBits  uint64 = 8
		platform     uint64 = 1
		server       uint64 = 2
		sysType      uint64 = 1
		seq          uint64 = 0
	)
	idGen := NewIdgen(platformBits, platform,
		serverBits, server,
		sysTypeBits, sysType,
		seq)
	go idGen.Recv()
	newId := idGen.GetNewId()
	assert.Equal(t, 2260613086576641, newId)
	assert.Equal(t, idGen.GetPlatform(2260613086576641), platform)
	assert.Equal(t, idGen.GetServer(2260613086576641), 2)
	assert.Equal(t, idGen.GetSysType(2260613086576641), 1)
}

//
func TestIdGen(t *testing.T) {
	id := NewDefaultIdgen(2, 1, 0, 0)
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
