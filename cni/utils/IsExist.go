package utils

func IsExistString(x string, y []string) bool {
	for i :=0; i < len(y); i++ {
		if y[i] == x {
			return true
		}
	}
	return false
}
