package collection

// MyString 是自定义的字符串类型
type MyString struct {
	value string
}

// NewMyString 构造函数
func NewMyString(val string) MyString {
	return MyString{value: val}
}

// Value 获取字符串值
func (s MyString) Value() string {
	return s.value
}
