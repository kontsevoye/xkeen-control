package xkeenipc

import "github.com/kontsevoye/xkeen-control/internal/executor"

func New(e executor.Executor) *XkeenIpc {
	return &XkeenIpc{e}
}

type XkeenIpc struct {
	e executor.Executor
}

func (i *XkeenIpc) Restart() error {
	return i.e.RunCommand("xkeen", "-restart")
}
