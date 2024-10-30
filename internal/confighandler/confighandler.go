package confighandler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/tailscale/hujson"
	"io"
	"os"
	"path/filepath"
	"time"
)

func New(filename string, enableBackups bool) *ConfigHandler {
	return &ConfigHandler{filename, enableBackups}
}

type ConfigHandler struct {
	filename      string
	enableBackups bool
}

type xrayRoutingConfig struct {
	Routing struct {
		Rules []rule `json:"rules"`
	} `json:"routing"`
}

type rule struct {
	InboundTag  []string `json:"inboundTag"`
	OutboundTag string   `json:"outboundTag"`
	Type        string   `json:"type"`
	Network     string   `json:"network,omitempty"`
	Port        string   `json:"port,omitempty"`
	Domains     []string `json:"domain,omitempty"`
	Ips         []string `json:"ip,omitempty"`
}

func (r *rule) isDomainsRule() bool {
	return r.OutboundTag == "vless-reality" && r.Type == "field" && r.Domains != nil
}

func (r *rule) hasDomain(domain string) bool {
	for _, d := range r.Domains {
		if d == domain {
			return true
		}
	}

	return false
}

func (ch *ConfigHandler) standardizeJSON(input []byte) ([]byte, error) {
	ast, err := hujson.Parse(input)
	if err != nil {
		return input, err
	}
	ast.Standardize()
	return ast.Pack(), nil
}

func (ch *ConfigHandler) loadConfig() (*xrayRoutingConfig, error) {
	data, err := os.ReadFile(ch.filename)
	if err != nil {
		return nil, err
	}

	standardizedData, err := ch.standardizeJSON(data)
	if err != nil {
		return nil, err
	}
	var c xrayRoutingConfig
	err = json.Unmarshal(standardizedData, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (ch *ConfigHandler) createBackupFile() error {
	if !ch.enableBackups {
		return nil
	}
	file, err := os.Open(ch.filename)
	if err != nil {
		return fmt.Errorf("ошибка открытия исходного файла: %w", err)
	}
	defer file.Close()

	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("ошибка генерации uuid: %w", err)
	}
	backupFile, err := os.Create(fmt.Sprintf("%s_%d_%s.bak", ch.filename, time.Now().Unix(), id.String()))
	if err != nil {
		return fmt.Errorf("ошибка создания файла назначения: %w", err)
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, file)
	if err != nil {
		return fmt.Errorf("ошибка копирования данных: %w", err)
	}

	err = backupFile.Sync()
	if err != nil {
		return fmt.Errorf("ошибка синхронизации файла: %w", err)
	}

	return nil
}

func (ch *ConfigHandler) saveConfig(config *xrayRoutingConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = ch.createBackupFile()
	if err != nil {
		return err
	}

	err = os.WriteFile(ch.filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (ch *ConfigHandler) GetDomains() ([]string, error) {
	c, err := ch.loadConfig()
	if err != nil {
		return []string{}, err
	}

	for _, rule := range c.Routing.Rules {
		if rule.isDomainsRule() {
			return rule.Domains, nil
		}
	}

	return []string{}, nil
}

func (ch *ConfigHandler) AddDomain(domain string) error {
	c, err := ch.loadConfig()
	if err != nil {
		return err
	}
	for i, rule := range c.Routing.Rules {
		if rule.isDomainsRule() {
			if rule.hasDomain(domain) {
				return fmt.Errorf("домен %s уже существует", domain)
			}
			c.Routing.Rules[i].Domains = append(rule.Domains, domain)
			err = ch.saveConfig(c)
			return err
		}
	}

	return nil
}

func (ch *ConfigHandler) DeleteDomain(domain string) error {
	c, err := ch.loadConfig()
	if err != nil {
		return err
	}
	for i, rule := range c.Routing.Rules {
		if rule.isDomainsRule() {
			for j, d := range rule.Domains {
				if d == domain {
					c.Routing.Rules[i].Domains = append(rule.Domains[:j], rule.Domains[j+1:]...)
					err = ch.saveConfig(c)
					return err
				}
			}
		}
	}

	return fmt.Errorf("домен %s не обнаружен и не был удален", domain)
}

func (ch *ConfigHandler) ListBackupFiles() ([]string, error) {
	backupFiles, err := filepath.Glob(fmt.Sprintf("%s_*.bak", ch.filename))
	if err != nil {
		return []string{}, err
	}

	return backupFiles, nil
}

func (ch *ConfigHandler) RestoreBackup(backupFileName string) error {
	backupFiles, err := ch.ListBackupFiles()
	if err != nil {
		return err
	}
	fileExists := false
	for _, backupFile := range backupFiles {
		if backupFile == backupFileName {
			fileExists = true
			break
		}
	}
	if !fileExists {
		return fmt.Errorf("файл для восстановления \"%s\" не существует", backupFileName)
	}
	backupFile, err := os.ReadFile(backupFileName)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл для восстановления \"%s\": %w", backupFileName, err)
	}
	err = ch.createBackupFile()
	if err != nil {
		return fmt.Errorf("не удалось создать бэкап текущей конфигурации: %w", err)
	}
	err = os.WriteFile(ch.filename, backupFile, 0644)
	if err != nil {
		return fmt.Errorf(
			"не удалось записать файл для восстановления \"%s\" в файл \"%s\": %w",
			backupFileName,
			ch.filename,
			err,
		)
	}

	return nil
}
