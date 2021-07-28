package utils

// MakeRange 创建一个连续数字的数组
func MakeRange(minNum int, mixNum int) []int {
	j := make([]int, mixNum-minNum+1)
	for i := range j {
		j[i] = minNum + i
	}
	return j
}

// Difference 求差集必须在其中一个数组是并集的基础上
func Difference(slice1 []string, slice2 []string) []string {
	//a是并集
	a := slice1
	b := slice2
	if len(a) > len(b) {
		a, b = b, a
	}
	m := make(map[string]int)
	n := make([]string, 0)
	for _, v := range a {
		m[v]++
	}
	for _, value := range b {
		times, _ := m[value]
		if times == 0 {
			n = append(n, value)
		}
	}
	return n
}
