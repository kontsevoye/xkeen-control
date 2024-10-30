package confighandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func createTmpFile() (*os.File, error) {
	tmpfile, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		return nil, err
	}

	testConfigFile, err := os.ReadFile("../../test/confighandler/config.json")
	if err != nil {
		return nil, err
	}
	testConfigData := string(testConfigFile)

	if _, err := tmpfile.Write([]byte(testConfigData)); err != nil {
		return nil, err
	}

	if err := tmpfile.Close(); err != nil {
		return nil, err
	}

	return tmpfile, nil
}

func TestGetDomains(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	domains, err := New(tmpFile.Name(), false).GetDomains()
	if err != nil {
		t.Fatalf("Ошибка получения доменов: %v", err)
	}

	expectedCount := 7
	if len(domains) != expectedCount {
		t.Errorf("Ожидалось %d доменов, получено: %d", expectedCount, len(domains))
	}

	expectedDomain := "ext:geosite_v2fly.dat:zoom"
	if domains[1] != expectedDomain {
		t.Errorf("Ожидалось доменное имя %s, получено: %s", expectedDomain, domains[1])
	}
}

func TestAddDomain(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	newDomain := "newdomain.com"
	ch := New(tmpFile.Name(), false)
	err = ch.AddDomain(newDomain)
	if err != nil {
		t.Fatalf("Ошибка при добавлении домена: %v", err)
	}

	domains, err := ch.GetDomains()
	if err != nil {
		t.Fatalf("Ошибка получения доменов: %v", err)
	}
	if domains[len(domains)-1] != newDomain {
		t.Errorf("Ожидалось доменное имя %s, получено: %s", newDomain, domains[len(domains)-1])
	}
}

func TestAddExistingDomain(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	newDomain := "ext:geosite_v2fly.dat:jetbrains"
	err = New(tmpFile.Name(), false).AddDomain(newDomain)
	if err == nil {
		t.Fatalf("Домен добавился без ошибки")
	}
	if err.Error() != fmt.Sprintf("домен %s уже существует", newDomain) {
		t.Fatalf("Непредвиденная ошибка: %v", err)
	}
}

func TestDeleteDomain(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	ch := New(tmpFile.Name(), false)
	domainsBefore, err := ch.GetDomains()
	if err != nil {
		t.Fatalf("Ошибка получения доменов: %v", err)
	}

	domainToDelete := "ext:geosite_v2fly.dat:jetbrains-ai"
	err = ch.DeleteDomain(domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	domainsAfter, err := ch.GetDomains()
	if err != nil {
		t.Fatalf("Ошибка получения доменов: %v", err)
	}
	if len(domainsAfter) == len(domainsBefore) {
		t.Fatalf("Список остался неизменным")
	}
	for _, domain := range domainsAfter {
		if domain == domainToDelete {
			t.Errorf("Домен %s не был удален", domainToDelete)
		}
	}
}

func TestDeleteUnknownDomain(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	domainToDelete := "koka.kola"
	err = New(tmpFile.Name(), false).DeleteDomain(domainToDelete)
	if err == nil {
		t.Fatalf("Домен удалился без ошибки")
	}
	if err.Error() != fmt.Sprintf("домен %s не обнаружен и не был удален", domainToDelete) {
		t.Fatalf("Непредвиденная ошибка: %v", err)
	}
}

func TestWriteDontBrokeConfig(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	domainToDelete := "ext:geosite_v2fly.dat:jetbrains-ai"
	err = New(tmpFile.Name(), false).DeleteDomain(domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	tmpFileContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка чтения временного файла: %v", err)
	}

	expectedConfigFile, err := os.ReadFile("../../test/confighandler/minified_config_after_delete_jb_ai.json")
	if err != nil {
		t.Fatalf("Ошибка чтения временного файла: %v", err)
	}
	expectedContent := string(expectedConfigFile)
	var compactedActualJSON bytes.Buffer
	err = json.Compact(&compactedActualJSON, tmpFileContent)
	if err != nil {
		t.Fatalf("Ошибка минификации JSON: %v", err)
	}
	actualContent := compactedActualJSON.String()
	if actualContent != expectedContent {
		t.Fatalf(
			"Содержимое не соответствует ожидаемому. Ожидали:\n\"%s\"\nПолучили:\n\"%s\"",
			expectedContent,
			actualContent,
		)
	}
}

func TestListBackupFiles(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	ch := New(tmpFile.Name(), true)
	backupFilesBefore, err := ch.ListBackupFiles()
	if err != nil {
		t.Fatalf("Ошибка получения бэкапов: %v", err)
	}
	if len(backupFilesBefore) != 0 {
		t.Fatalf("Количество бэкапов в начале ожидалось 0, получилось %d", len(backupFilesBefore))
	}

	domainToDelete := "ext:geosite_v2fly.dat:jetbrains-ai"
	err = ch.DeleteDomain(domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	backupFilesAfter, err := ch.ListBackupFiles()
	if err != nil {
		t.Fatalf("Ошибка получения бэкапов: %v", err)
	}
	if len(backupFilesAfter) != 1 {
		t.Fatalf("Количество бэкапов в конце ожидалось 1, получилось %d", len(backupFilesAfter))
	}
}

func TestRestoreBackup(t *testing.T) {
	tmpFile, err := createTmpFile()
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	domainToDelete := "ext:geosite_v2fly.dat:jetbrains-ai"
	ch := New(tmpFile.Name(), true)
	err = ch.DeleteDomain(domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	backupFiles, err := ch.ListBackupFiles()
	if err != nil {
		t.Fatalf("Ошибка получения бэкапов: %v", err)
	}
	backupFileName := backupFiles[len(backupFiles)-1]
	err = ch.RestoreBackup(backupFileName)
	if err != nil {
		t.Fatalf("Ошибка восстановления бэкапа: %v", err)
	}

	actualFile, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка чтения файла: %v", err)
	}
	actualFileContent := string(actualFile)
	expectedFile, err := os.ReadFile(backupFileName)
	if err != nil {
		t.Fatalf("Ошибка чтения временного файла: %v", err)
	}
	expectedFileContent := string(expectedFile)

	if actualFileContent != expectedFileContent {
		t.Fatalf(
			"Файл после восстановления не соответствует ожианиям. Ожидалось:\n%s\nПолучилось:\n%s",
			expectedFileContent,
			actualFileContent,
		)
	}
}
