package xkeenipc

import "github.com/kontsevoye/xkeen-control/internal/executor"

func New(e executor.Executor, binaryName string) *XkeenIpc {
	if binaryName == "" {
		binaryName = "xkeen"
	}
	return &XkeenIpc{e, binaryName}
}

type XkeenIpc struct {
	e          executor.Executor
	binaryName string
}

func (i *XkeenIpc) Restart() error {
	return i.e.RunCommand(i.binaryName, "-restart")
}
