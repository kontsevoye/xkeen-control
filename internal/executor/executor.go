package executor

import (
	"fmt"
	"os/exec"
)

func NewExecExecutor() *ExecExecutor {
	return &ExecExecutor{}
}

type Executor interface {
	RunCommand(name string, arg ...string) error
}

type ExecExecutor struct {
}

func (e *ExecExecutor) RunCommand(name string, arg ...string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("бинарник %s не обнаружен", name)
	}

	return exec.Command(name, arg...).Run()
}
