package idgen

import (
	"fmt"
	"runtime"
	"time"
)

type IdGen struct {
	platformBits uint64 // if =12 , 最多可以有2^12个平台 4096
	serverBits   uint64 // if =9  最多可以有2^9个服务器 512
	sysTypeBits  uint64 //  特殊标记位,一些特殊的系统可能需要 如果=3，sysType 最多有2^3个
	platform     uint64 //平台
	server       uint64 // server 编号
	sysType      uint64 // 额外提供的一个字段，跟platform,server 字段一样，也占据n个字节
	sequence     uint64 // 自增系列
	ch           chan chan uint64
	isRunning    bool
	stopChan     chan bool
}

// golang mysql driver 不支持uint64 最高位设成1，
// 所以 所有移位操作，  以63 为最高位数
func (idGen IdGen) GetSeqBits() uint64 {
	return 63 - idGen.platformBits - idGen.serverBits - idGen.sysTypeBits
}
func (idGen IdGen) GetMaxSequence() uint64 {
	// math.power(2,4)=16, 2<<3=16
	return 2 << (idGen.GetSeqBits() - 1) // 2<<17 == 2^18
}
func (idGen IdGen) GetSequenceMask() uint64 {
	return ((1 << idGen.GetSeqBits()) - 1)
}

func (idGen IdGen) GetPlatformShift() uint64 {
	return (63 - idGen.platformBits)
}
func (idGen IdGen) GetPlatformMask() uint64 {
	return ((1 << idGen.platformBits) - 1) << idGen.GetPlatformShift()
}

func (idGen IdGen) GetServerShift() uint64 {
	return 63 - (idGen.platformBits + idGen.serverBits)
}
func (idGen IdGen) GetServerMask() uint64 {
	return ((1 << idGen.serverBits) - 1) << idGen.GetServerShift()
}

func (idGen IdGen) GetSysTypeShift() uint64 {
	return 63 - (idGen.platformBits + idGen.serverBits + idGen.sysTypeBits)
}

func (idGen IdGen) GetSysTypeMask() uint64 {
	return ((1 << idGen.sysTypeBits) - 1) << idGen.GetSysTypeShift()
}

func (idGen IdGen) GetPlatform(id uint64) uint64 {
	return id & idGen.GetPlatformMask() >> idGen.GetPlatformShift()
}
func (idGen IdGen) GetServer(id uint64) uint64 {
	return id & idGen.GetServerMask() >> idGen.GetServerShift()
}
func (idGen IdGen) GetSysType(id uint64) uint64 {
	return id & idGen.GetSysTypeMask() >> idGen.GetSysTypeShift()
}

func (idGen IdGen) GetIdSequence(id uint64) uint64 {
	return id & idGen.GetSequenceMask()
}
func (idGen *IdGen) SetSequence(sequence uint64) {
	idGen.sequence = sequence
}

func NewIdgen(platformBits uint64, platform uint64, serverBits uint64, server uint64, sysTypeBits uint64, systype uint64, sequence uint64) (idgen *IdGen) {
	return &IdGen{
		platformBits: platformBits,
		serverBits:   serverBits,
		sysTypeBits:  sysTypeBits,
		platform:     platform,
		server:       server,
		sysType:      systype,
		sequence:     sequence,
		ch:           make(chan chan uint64),
		stopChan:     make(chan bool),
	}
}

func NewDefaultIdgen(platform uint64, server uint64, systype uint64, sequence uint64) (idgen *IdGen) {
	return NewIdgen(12, platform, 9, server, 8, systype, sequence)
}

func (idGen *IdGen) newId(systype uint64) (newId uint64) {
	idGen.sequence = idGen.sequence + 1
	newId = (idGen.platform << idGen.GetPlatformShift()) | (idGen.server << idGen.GetServerShift()) | idGen.sysType<<idGen.GetSysTypeShift() | (idGen.sequence)
	return
}
func (idGen *IdGen) Stop() {
	if idGen.isRunning == false {
		return
	}
	select {
	case idGen.stopChan <- true:
	case <-time.After(time.Second):
		fmt.Println("stop idgen timeout")
	}
	idGen.isRunning = false
}
func (idGen *IdGen) Recv() {
	protectFunc(
		func() {
			idGen.isRunning = true
			for idGen.isRunning {
				select {
				case container := <-idGen.ch:
					container <- idGen.newId(idGen.sysType)
				case <-idGen.stopChan:
					idGen.isRunning = false
				}
			}
		})
}

// 应该返回int64
func (idGen *IdGen) GetNewId() uint64 {
	container := make(chan uint64)
	idGen.ch <- container
	id := <-container
	close(container)
	return id
}

func protectFunc(fun func()) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
			for i := 0; i < 10; i++ {
				funcName, file, line, ok := runtime.Caller(i)
				if ok {
					fmt.Printf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
				}
			}
		}
	}()
	fun()
}
