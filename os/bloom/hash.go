package bloom

// BKDRHash 使用 BKDR 哈希算法计算字符串的哈希值
// 参数：
//   - str: 输入字符串
//
// 返回：
//   - uint64: 计算得到的哈希值
func BKDRHash(str string) uint64 {
	const seed uint64 = 131 // 31 131 1313 13131 131313 等都是常用的种子值
	var hash uint64

	for i := 0; i < len(str); i++ {
		hash = (hash * seed) + uint64(str[i])
	}

	return hash & 0x7FFFFFFF
}

// SDBMHash 使用 SDBM 哈希算法计算字符串的哈希值
// 参数：
//   - str: 输入字符串
//
// 返回：
//   - uint64: 计算得到的哈希值
func SDBMHash(str string) uint64 {
	var hash uint64

	for i := 0; i < len(str); i++ {
		hash = uint64(str[i]) + (hash << 6) + (hash << 16) - hash
	}

	return hash & 0x7FFFFFFF
}

// DJBHash 使用 DJB 哈希算法计算字符串的哈希值
// 参数：
//   - str: 输入字符串
//
// 返回：
//   - uint64: 计算得到的哈希值
func DJBHash(str string) uint64 {
	var hash uint64

	for i := 0; i < len(str); i++ {
		hash = ((hash << 5) + hash) + uint64(str[i])
	}

	return hash & 0x7FFFFFFF
}
