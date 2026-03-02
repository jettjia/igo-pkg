package php2go

// Return
//
//	@Description: if实现的三元表达式
//	@param boolExpression: 布尔表达式，最终返回一个布尔值
//	@param trueReturnValue: 当boolExpression返回值为true的时候返回的值
//	@param falseReturnValue: 当boolExpression返回值为false的时候返回的值
//	@return bool: 三元表达式的结果，为trueReturnValue或者falseReturnValue中的一个
func IfExpression[T any](boolExpression bool, trueReturnValue, falseReturnValue T) T {
	if boolExpression {
		return trueReturnValue
	} else {
		return falseReturnValue
	}
}
