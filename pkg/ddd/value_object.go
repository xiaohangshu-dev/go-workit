package ddd

// IValueObject 值对象接口，值对象通过其属性值来定义相等性。
type IValueObject interface {
	// Equal 比较两个值对象是否相等
	Equal(other IValueObject) bool
}

// ValueObject 值对象基类，嵌入此类型表示该结构体是一个值对象。
// 注意：Go 的嵌入类型无法访问外层结构体的字段，
// 因此 Equal 方法需要在具体的值对象类型中自行实现。
// 此基类仅作为标记，表明该类型是一个值对象。
type ValueObject struct{}
