package utils

func IsExistString(str string, strSlice []string) bool {
	for i := 0; i < len(strSlice); i++ {
		if strSlice[i] == str {
			return true
		}
	}
	return false
}

func IsExistByte(byte byte, byteSlice []byte) bool {
	for i := 0; i < len(byteSlice); i++ {
		if byteSlice[i] == byte {
			return true
		}
	}
	return false
}
