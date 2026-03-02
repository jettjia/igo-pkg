package enum

type EnumType int

const (
	CONSOLE EnumType = iota + 1 // define enumeration type constants based on iota characteristics and output them to the console
	FILE
	MQ
	ES
	ZS
	OTEL
)

var weekdayStr = []string{"console", "file", "mq", "es", "zs", "otel"}

// String - returns the index value of the enumerated item
func (w EnumType) String() string {
	return weekdayStr[w-1]
}

// Index - returns the character value of the enumerated item
func (w EnumType) Index() int {
	return int(w)
}

// Values returns all values of the enumeration
func Values() []string {
	return weekdayStr
}

// ExistOf determine whether a value exists in the enumerated value
func ExistOf(str string) bool {
	for _, v := range weekdayStr {
		if v == str {
			return true
		}
	}
	return false
}
