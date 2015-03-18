package idgen

import (
	"fmt"
	"time"
)

// 1年=365*24*3600*1000ms=2#111 01010111 10110001 00101100 00000000 2#35位
// 1年=365*24*3600s=2#111 2#111 1000010 0110011 10000000  2#27位
// 10年=             ms     2#100 1001011 0110011 1010111 0111000 00000000 2#43位
// 10年=             s    1:  2# 1001011 0011000 0000011 00000000  2#32位

const (
	// 所存储的时间是一个时间差是 当前时间与baseTimeStamp 时间差对应的秒
	// 你用的时候可以调整baseTimeStamp到当前时间 ，但是正式使用之后就不要调整此值了，否则新生成的
	// id可能会与老id冲突
	BASE_TIME_STAMP = "2015-03-17 22:45:59" // 格式里的0不能少， 比如03 不能写成3 否则解析会出错
	// 这里用32位来存储时间戳相关信息， 生成的id 10年内不会有冲突，如果生成的id 有可能10年后依然在使用，则需要相应调大此值
	// 如果仅使用1年，则27位就够了， 可以根据需要 适当调整此数值
	TIME_STAMP_BITS = 32 // 存到秒级
	// uint64除最高位不用，及timestampBits外 ，63-32=31 位， 还有空余的31位可用
	// 假如1秒内最大有可能生成1024个id 的话， 则用于自增序列sequence 所占的位数就是10
	// 所以  platformBit+serverBits+sysTypeBits=最大有31-10 =21位可用
	// 您需要适当调整这3个值所占的位数
	//
)

type IdGen3 struct {
	idGenBaseTimestamp uint64
	platformBits       uint64 // if =12 , 最多可以有2^12个平台 4096
	serverBits         uint64 // if =9  最多可以有2^9个服务器 512
	sysTypeBits        uint64 //  特殊标记位,一些特殊的系统可能需要 如果=3，sysType 最多有2^3个
	Platform           uint64 //平台
	Server             uint64 // Server 编号
	SysType            uint64 // 额外提供的一个字段，跟platform,Server 字段一样，也占据n个字节
	timeStamp          uint64

	sequence  uint64 // 自增系列
	ch        chan chan uint64
	isRunning bool
	stopChan  chan bool
	maxSeq    uint64
}

func (this IdGen3) GetPlatformShift() uint64 {
	return (63 - this.platformBits)
}
func (this IdGen3) GetPlatformMask() uint64 {
	return ((1 << this.platformBits) - 1) << this.GetPlatformShift()
}

func (this IdGen3) GetServerShift() uint64 {
	return 63 - (this.platformBits + this.serverBits)
}
func (this IdGen3) GetServerMask() uint64 {
	return ((1 << this.serverBits) - 1) << this.GetServerShift()
}

func (this IdGen3) GetSysTypeShift() uint64 {
	return 63 - (this.platformBits + this.serverBits + this.sysTypeBits)
}

func (this IdGen3) GetSysTypeMask() uint64 {
	return ((1 << this.sysTypeBits) - 1) << this.GetSysTypeShift()
}
func (this IdGen3) GetTimeStampShift() uint64 {
	return 63 - (this.platformBits + this.serverBits + this.sysTypeBits + TIME_STAMP_BITS)
}
func (this IdGen3) GetTimeStampMask() uint64 {
	return ((1 << TIME_STAMP_BITS) - 1) << this.GetTimeStampShift()
}

// golang mysql driver 不支持uint64 最高位设成1，
// 所以 所有移位操作，  以63 为最高位数
func (this IdGen3) GetSeqBits() uint64 {
	return 63 - this.platformBits - this.serverBits - this.sysTypeBits - TIME_STAMP_BITS
}
func (this IdGen3) GetMaxSequence() uint64 {
	// math.power(2,4)=16, 2<<3=16
	return 2 << (this.GetSeqBits() - 1) // 2<<17 == 2^18
}
func (this IdGen3) GetSequenceMask() uint64 {
	return ((1 << this.GetSeqBits()) - 1)
}

func (this IdGen3) GetPlatform(id uint64) uint64 {
	return id & this.GetPlatformMask() >> this.GetPlatformShift()
}
func (this IdGen3) GetServer(id uint64) uint64 {
	return id & this.GetServerMask() >> this.GetServerShift()
}
func (this IdGen3) GetSysType(id uint64) uint64 {
	return id & this.GetSysTypeMask() >> this.GetSysTypeShift()
}

func (this IdGen3) GetIdSequence(id uint64) uint64 {
	return id & this.GetSequenceMask()
}
func (this *IdGen3) SetSequence(sequence uint64) {
	this.sequence = sequence
}

func (this *IdGen3) cleanSequence() {
	currentT := uint64(time.Now().UnixNano() / 1000000000)
	if currentT != this.timeStamp {
		this.timeStamp = currentT
		this.sequence = 0
	}
}

func (this *IdGen3) addTimeStamp() {
	currentT := uint64(time.Now().UnixNano() / 1000000000)
	if currentT < this.timeStamp {
		fmt.Println("[Warning],id生成器sequence所占位数过少，1秒内生成的id超过其最大值，下次id生成器启动时生成的id 有可能会与现在的id 冲突，因为现在id生成器使用的时间戳已经超过当前时间,已超过当前时间多少秒:", this.timeStamp-currentT)
	}

	this.timeStamp += 1
	this.sequence = 0
}
func (this *IdGen3) recv() {
	for {
		select {
		case container := <-this.ch:
			this.sequence = this.sequence + 1
			if this.sequence >= this.maxSeq {
				this.addTimeStamp()
			}
			newId := (this.Platform << this.GetPlatformShift()) | (this.Server << this.GetServerShift()) | this.SysType<<this.GetSysTypeShift() | (this.timeStamp-this.idGenBaseTimestamp)<<this.GetTimeStampShift() | (this.sequence)
			container <- newId

		}
	}
}

func NewIdgen3(platformBits, Platform, serverBits, Server, sysTypeBits, systype uint64) (idgen *IdGen3) {
	idGenBaseTimestamp, err := time.Parse("2006-01-02 15:04:05", BASE_TIME_STAMP)
	if err != nil {
		fmt.Println("err 解析baseTimeStamp发生错误 ，时间格式不对", BASE_TIME_STAMP, "应该形如2006-01-02 15:04:05")
		panic("err 解析baseTimeStamp发生错误")
	}

	timeStamp := time.Now().UnixNano() / 1000000000

	idgen = &IdGen3{
		platformBits:       platformBits,
		serverBits:         serverBits,
		sysTypeBits:        sysTypeBits,
		Platform:           Platform,
		Server:             Server,
		SysType:            systype,
		timeStamp:          uint64(timeStamp),
		sequence:           0,
		idGenBaseTimestamp: uint64(idGenBaseTimestamp.UnixNano() / 1000000000),
		ch:                 make(chan chan uint64),
	}
	fmt.Printf("platformBits=%d,serverBits=%d,sysTypeBits=%d,seqBits=%d,timeStampBits=%d,maxSeq/seconds=%d\n",
		idgen.platformBits, idgen.serverBits, idgen.sysTypeBits, idgen.GetSeqBits(), TIME_STAMP_BITS, idgen.GetMaxSequence())
	if idgen.GetSeqBits() < 10 {
		panic("[Error]idgen.SeqBits<10,maybe not each for each second 2^10=1024")
	}
	idgen.maxSeq = idgen.GetMaxSequence()
	go idgen.recv()
	return
}

func NewDefaultIdgen3(Platform, Server, systype uint64) (idgen *IdGen3) {
	// 相加最好不要超过15 ，63-32-15=16  即seq 1秒内的最大值 2^16=65536
	idgen = NewIdgen3(
		4, Platform,
		8, Server,
		4, systype)
	return
}

func (this *IdGen3) GetNewId() uint64 {
	container := make(chan uint64)
	this.ch <- container
	id := <-container
	close(container)
	return id

}
