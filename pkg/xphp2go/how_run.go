package php2go

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------、

type RunType int

const (

	// RunTypeUnknown 咱也不知道咋运行的
	RunTypeUnknown RunType = iota

	// RunTypeSourceCode 是从源代码中运行的
	RunTypeSourceCode

	// RunTypeReleaseBinary 发布的二进制文件运行
	RunTypeReleaseBinary
)

func (x RunType) String() string {
	switch x {
	case RunTypeUnknown:
		return "Unknown"
	case RunTypeSourceCode:
		return "SourceCode"
	case RunTypeReleaseBinary:
		return "ReleaseBinary"
	default:
		return ""
	}
}

// GetRunType 获取当前程序是以什么方式运行的
func GetRunType() (RunType, error) {
	executable, err := os.Executable()
	if err != nil {
		return RunTypeUnknown, err
	}
	switch runtime.GOOS {
	case "windows":
		// C:\Users\5950X\AppData\Local\Temp\GoLand\___go_build_github_com_golang_infrastructure_go_project_root_directory_main_test.exe
		if strings.Contains(executable, "\\AppData\\Local\\Temp\\") {
			return RunTypeSourceCode, nil
		} else {
			return RunTypeReleaseBinary, nil
		}
	case "linux":
		// /tmp/go-build1325605723/b001/exe/test
		if strings.HasPrefix(executable, "/tmp/go-build") {
			return RunTypeSourceCode, nil
		} else {
			return RunTypeReleaseBinary, nil
		}
	case "darwin":
		// /var/folders/kd/dzyx8fc96fx4j3mtdtjsl4z40000gn/T/go-build3362823274/b001/exe/main
		if strings.HasSuffix(executable, "/exe/main") && strings.Contains(executable, "/go-build") {
			return RunTypeSourceCode, nil
		} else {
			return RunTypeReleaseBinary, nil
		}
	}
	return RunTypeUnknown, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

// SourceCodeRunType 如果是从源代码运行的话，则入口是啥
type SourceCodeRunType int

const (

	// SourceCodeRunTypeUnknown 只知道是从源代码中运行的，但是入口俺也不知道
	SourceCodeRunTypeUnknown SourceCodeRunType = iota

	// SourceCodeRunTypeExample 是从Example运行的
	SourceCodeRunTypeExample

	// SourceCodeRunTypeBenchmark 是从Benchmark运行的
	SourceCodeRunTypeBenchmark

	// SourceCodeRunTypeMain 是从main.go运行的
	SourceCodeRunTypeMain

	// SourceCodeRunTypeTest 是从测试用例运行的
	SourceCodeRunTypeTest
)

func (x SourceCodeRunType) String() string {
	switch x {
	case SourceCodeRunTypeUnknown:
		return "Unknown"
	case SourceCodeRunTypeExample:
		return "Example"
	case SourceCodeRunTypeBenchmark:
		return "Benchmark"
	case SourceCodeRunTypeMain:
		return "main.go"
	case SourceCodeRunTypeTest:
		return "Test"
	default:
		return ""
	}
}

// GetSourceCodeRunType 如果是从源代码运行的，则是以什么方式运行的
func GetSourceCodeRunType() (SourceCodeRunType, error) {
	stack := string(debug.Stack())
	split := strings.Split(stack, "\n")
	if len(split) < 3 {
		return SourceCodeRunTypeUnknown, nil
	}
	s := split[len(split)-3]
	switch s {
	case "created by testing.(*T).Run":
		return SourceCodeRunTypeTest, nil
	case "created by testing.(*B).run1":
		return SourceCodeRunTypeBenchmark, nil
	case "main.main()":
		if strings.HasPrefix(split[len(split)-2], "\t_testmain.go") {
			return SourceCodeRunTypeExample, nil
		} else {
			return SourceCodeRunTypeMain, nil
		}
	}
	return SourceCodeRunTypeUnknown, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

// RunIDE 是从什么IDE中运行的
type RunIDE int

const (

	// RunIDEUnknown 咱也不知道是啥IDE
	RunIDEUnknown RunIDE = iota

	// RunIDEGoland 是从GoLand中运行的
	RunIDEGoland
)

func (x RunIDE) String() string {
	switch x {
	case RunIDEUnknown:
		return "Unknown"
	case RunIDEGoland:
		return "Goland"
	default:
		return ""
	}
}

func GetRunIDE() (RunIDE, error) {
	executable, err := os.Executable()
	if err != nil {
		return RunIDEUnknown, err
	}
	dir := filepath.Dir(executable)
	switch runtime.GOOS {
	case "windows":
		if strings.HasSuffix(dir, "\\Temp\\GoLand") {
			return RunIDEGoland, nil
		}
	default:
		if strings.HasSuffix(dir, "GoLand") {
			return RunIDEGoland, nil
		}
	}
	return RunIDEUnknown, nil
}

// ------------------------------------------------- --------------------------------------------------------------------
