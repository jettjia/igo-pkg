package dm

import (
	"fmt"
	"testing"
)

func Test_MysqlType2Dm(t *testing.T) {
	s1, err := MysqlType2Dm("bigint")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(s1)
}

func Test_RemoveParentheses(t *testing.T) {
	s1 := RemoveParentheses("bigint(20)")
	fmt.Println(s1)

	s2 := RemoveParentheses("bigint(20) unsigned")
	fmt.Println(s2)

	s3 := RemoveParentheses("int(20) unsigned")
	fmt.Println(s3)
}
