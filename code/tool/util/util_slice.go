package util

// AllNil 检查切片中的所有元素是否都为nil
func AllNil(slice interface{}) bool {
	// 获取切片的长度
	length := len(slice.([]interface{}))

	// 遍历切片，检查每个元素是否为nil
	for i := 0; i < length; i++ {
		if slice.([]interface{})[i] != nil {
			return false
		}
	}
	return true
}
