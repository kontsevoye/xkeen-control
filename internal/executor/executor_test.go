package executor

import "testing"

func TestRunCommandSuccess(t *testing.T) {
	e := NewExecExecutor()
	err := e.RunCommand("echo", "itworks")
	if err != nil {
		t.Fatalf("Не ожидаемая ошибка: %v", err)
	}
}

func TestRunCommandError(t *testing.T) {
	e := NewExecExecutor()
	err := e.RunCommand("zzzalupaaaa", "itdoesntworks")
	if err == nil {
		t.Fatalf("Запуск без ожидаемой ошибки")
	}
	if err.Error() != "бинарник zzzalupaaaa не обнаружен" {
		t.Fatalf("Ошибка не совпадает с ожидаемой: %v", err)
	}
}
