package php2go

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRunType(t *testing.T) {
	runType, err := GetRunType()
	assert.Nil(t, err)
	assert.Equal(t, RunTypeSourceCode, runType)
}

func TestGetSourceCodeRunType(t *testing.T) {
	runType, err := GetSourceCodeRunType()
	assert.Nil(t, err)
	assert.Equal(t, SourceCodeRunTypeTest, runType)
}

func ExampleGetSourceCodeRunType() {
	runType, err := GetSourceCodeRunType()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(runType)
	// Output:
	// 1
}

func BenchmarkGetSourceCodeRunType(b *testing.B) {
	runType, err := GetSourceCodeRunType()
	assert.Nil(b, err)
	assert.Equal(b, SourceCodeRunTypeBenchmark, runType)
}

func TestGetRunIDE(t *testing.T) {
	ide, err := GetRunIDE()
	assert.Nil(t, err)
	assert.Equal(t, ide, RunIDEGoland)
}
