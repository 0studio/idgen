package idgen

import (
	"fmt"
	"runtime"
	"time"
)

// http://mikespook.com/2012/05/golang-funny-play-with-channel/
// 不会有协程安全问题，因为都是通过channel来读取newId
//
type IdGen struct {
	platform uint64 //
	serverid uint64
	// lastTimestamp uint64
	sysType   uint64
	seq       uint64
	ch        chan chan uint64
	isRunning bool
	stopChan  chan bool
}

/*
* 最多可以有2^12个平台 4096
* 最多可以有2^11个服务器 2048
 */

const PlatformBits = 12 //

// func MakePlatformId(channelId int32, OS int32) uint64 {
// 	return uint64((channelId << OSBits) | OS)

// }

const ServeridBits = 9

// 特殊标记位,一些特殊的系统可能需要
const SysTypeBits = 8

// const TimestampBits = 28
// golang mysql driver 不支持uint64 最高位设成1，
// 所以 所有移位操作，  以63 为最高位数
const SeqBits = 63 - (PlatformBits + ServeridBits + SysTypeBits)

// // 2^18=262144
// const MaxSeq = 2 << (63 - SeqBits - 1) // 2<<17 == 2^18
// // seqbits = 63-pfbits-serveridbits-timestampbits

const PLATFORM_SHIFT = (63 - PlatformBits)
const SERVER_SHIFT = 63 - (PlatformBits + ServeridBits)
const SYSTYPE_SHIFT = 63 - (PlatformBits + ServeridBits + SysTypeBits)

// const TimestampShift = 63 - (PlatformBits + ServeridBits + TimestampBits)

const PLATFORM_MASK = ((1 << PlatformBits) - 1) << PLATFORM_SHIFT

func GetIdPlatformId(id uint64) uint64 {
	return id & PLATFORM_MASK >> PLATFORM_SHIFT
}

// func GetChannelId(id uint64) uint64 {
// 	return GetIdPlatformId(id) >> OSBits
// }
// func GetChannelIdFromPlatformId(platformid uint64) uint64 {
// 	return platformid >> OSBits
// }

// func GetOSId(id uint64) uint64 {
// 	return GetIdPlatformId(id) & OS_MASK
// }

const SERVER_MASK = ((1 << ServeridBits) - 1) << SERVER_SHIFT

func GetIdServerId(id uint64) uint64 {
	return id & SERVER_MASK >> SERVER_SHIFT
}

const SYSTYPE_MASK = ((1 << SysTypeBits) - 1) << SYSTYPE_SHIFT

func GetIdSysTypeId(id uint64) uint64 {
	return id & SYSTYPE_MASK >> SYSTYPE_SHIFT
}

const SEQ_MASK = ((1 << SeqBits) - 1)

func GetIdSeq(id uint64) uint64 {
	return id & SEQ_MASK
}

func MakeId(platform uint64, server uint64, systype uint64, seq uint64) uint64 {
	return (platform << PLATFORM_SHIFT) | (server << SERVER_SHIFT) | systype<<SYSTYPE_SHIFT | (seq)
}

func NewIdgen(platform uint64, serverid uint64, systype uint64, base uint64) (idgen *IdGen) {
	return &IdGen{
		platform: platform,
		serverid: serverid,
		sysType:  systype,
		seq:      base,
		ch:       make(chan chan uint64),
		stopChan: make(chan bool),
	}
}

func (this *IdGen) newId(systype uint64) (newId uint64) {
	this.seq = this.seq + 1

	newId = MakeId(this.platform, this.serverid, systype, this.seq)

	return
}
func (this *IdGen) Stop() {
	if this.isRunning == false {
		return
	}
	select {
	case this.stopChan <- true:
	case <-time.After(time.Second):
		fmt.Println("stop idgen timeout")
	}
	this.isRunning = false
}
func (this *IdGen) Recv() {
	protectFunc(
		func() {
			// defer utils.PrintPanicStack()
			// var newid uint64
			// var sysTypeId uint64
			this.isRunning = true
			for this.isRunning {
				// atomic.StoreInt64(&newid, this.newId())
				// 总感觉 算newid的时候 不是原子操作
				select {
				case container := <-this.ch:
					// select { // select 避免 从ch 读数据的时候 卡死在这 ，
					// case sysTypeId = <-container:
					// case <-time.After(30 * time.Millisecond): // 300ms 后， 如果依然没有从申请id的goroute那收到platform
					// 	fmt.Println("use default systype")
					// 	sysTypeId = this.sysType // 使用默认的platform
					// }
					container <- this.newId(this.sysType)

				case <-this.stopChan:
					this.isRunning = false
				}
			}
		})
}

// 应该返回int64
func (this *IdGen) GetNewId() uint64 {
	container := make(chan uint64)
	this.ch <- container
	// container <- systypeId
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
