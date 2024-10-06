package xkeenipc

import (
	"fmt"
	"os/exec"
)

func Restart() error {
	_, err := exec.LookPath("xkeen")
	if err != nil {
		return fmt.Errorf("бинарник xkeen не обнаружен")
	}

	return exec.Command("xkeen", "-restart").Run()
}
