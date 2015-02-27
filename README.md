# Demo
```
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
    idGen.SetSequence(100)// if needed
	go idGen.Recv()
	newId := idGen.GetNewId() //	assert.Equal(t, 2260613086576641, newId)
```
