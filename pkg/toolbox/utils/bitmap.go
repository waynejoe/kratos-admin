package utils

// 判断指定索引位是否为1（低位在前模式）
func IsBitSet(data []byte, index int32) bool {
	if index < 0 {
		return false
	}
	if int(index) >= len(data)*8 {
		return false
	}
	bytePos := index >> 3
	bitPos := index & 0x7
	return (data[bytePos] & (1 << bitPos)) != 0
}

func SetBit(data []byte, index int32, value bool) []byte {
	if index < 0 {
		return data
	}

	// 计算所需字节数 (index/8 +1)
	requiredBytes := (index >> 3) + 1

	// 动态扩容逻辑
	if len(data) < int(index) {
		newData := make([]byte, requiredBytes)
		copy(newData, data)
		data = newData
	}

	// 位操作（设置bit为1）
	bytePos := index >> 3
	bitPos := index & 0x7
	if value {
		data[bytePos] |= 1 << bitPos
	} else {
		data[bytePos] &^= 1 << bitPos
	}

	return data
}
