package xkeenipc

import (
	"fmt"
	"testing"
)

type mockExecutor struct {
	knownName string
	knownArgs []string
}

func (e *mockExecutor) RunCommand(name string, arg ...string) error {
	if name != e.knownName {
		return fmt.Errorf("бинарник %s не обнаружен", name)
	}

	if len(e.knownArgs) != len(arg) {
		return fmt.Errorf("несовпадение количества: %d/%d", len(e.knownArgs), len(arg))
	}

	for i, v := range e.knownArgs {
		if v != arg[i] {
			return fmt.Errorf("неизвестный аргумент: %s/%s", v, arg[i])
		}
	}

	return nil
}

func TestRestartUnknownBinary(t *testing.T) {
	e := New(&mockExecutor{knownName: "test", knownArgs: []string{}}, "")
	err := e.Restart()
	if err == nil {
		t.Fatalf("Перезапуск без ожидаемой ошибки")
	}
	if err.Error() != "бинарник xkeen не обнаружен" {
		t.Fatalf("Ошибка не совпадает с ожидаемой: %v", err)
	}
}

func TestRestartUnknownArg(t *testing.T) {
	e := New(&mockExecutor{knownName: "xkeen", knownArgs: []string{}}, "")
	err := e.Restart()
	if err == nil {
		t.Fatalf("Перезапуск без ожидаемой ошибки")
	}
	if err.Error() != "несовпадение количества: 0/1" {
		t.Fatalf("Ошибка не совпадает с ожидаемой: %v", err)
	}
}

func TestRestartSuccess(t *testing.T) {
	e := New(&mockExecutor{knownName: "xkeen", knownArgs: []string{"-restart"}}, "")
	err := e.Restart()
	if err != nil {
		t.Fatalf("Не ожидаемая ошибка: %v", err)
	}
}
