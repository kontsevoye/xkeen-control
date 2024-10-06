package confighandler

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func createTmpFile() (*os.File, error) {
	tmpfile, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		return nil, err
	}

	testConfigFile, err := os.ReadFile("../test_data/config.json")
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

	domains, err := GetDomains(tmpFile.Name())
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
	err = AddDomain(tmpFile.Name(), newDomain)
	if err != nil {
		t.Fatalf("Ошибка при добавлении домена: %v", err)
	}

	domains, err := GetDomains(tmpFile.Name())
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
	err = AddDomain(tmpFile.Name(), newDomain)
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

	domainsBefore, err := GetDomains(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка получения доменов: %v", err)
	}

	domainToDelete := "ext:geosite_v2fly.dat:jetbrains-ai"
	err = DeleteDomain(tmpFile.Name(), domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	domainsAfter, err := GetDomains(tmpFile.Name())
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
	err = DeleteDomain(tmpFile.Name(), domainToDelete)
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
	err = DeleteDomain(tmpFile.Name(), domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	tmpFileContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка чтения временного файла: %v", err)
	}

	expectedConfigFile, err := os.ReadFile("../test_data/config_after_delete_jb_ai.json")
	if err != nil {
		t.Fatalf("Ошибка чтения временного файла: %v", err)
	}
	expectedContent := string(expectedConfigFile)
	expectedContent = strings.Trim(expectedContent, " ")
	expectedContent = strings.Trim(expectedContent, "\n")
	actualContent := string(tmpFileContent)
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

	backupFilesBefore, err := ListBackupFiles(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка получения бэкапов: %v", err)
	}
	if len(backupFilesBefore) != 0 {
		t.Fatalf("Количество бэкапов в начале ожидалось 0, получилось %d", len(backupFilesBefore))
	}

	domainToDelete := "ext:geosite_v2fly.dat:jetbrains-ai"
	err = DeleteDomain(tmpFile.Name(), domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	backupFilesAfter, err := ListBackupFiles(tmpFile.Name())
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
	err = DeleteDomain(tmpFile.Name(), domainToDelete)
	if err != nil {
		t.Fatalf("Ошибка удаления домена: %v", err)
	}

	backupFiles, err := ListBackupFiles(tmpFile.Name())
	if err != nil {
		t.Fatalf("Ошибка получения бэкапов: %v", err)
	}
	backupFileName := backupFiles[len(backupFiles)-1]
	err = RestoreBackup(tmpFile.Name(), backupFileName)
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
