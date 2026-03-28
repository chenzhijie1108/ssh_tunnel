package main

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Tunnels []TunnelConfig `json:"tunnels"`
}

type TunnelConfig struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	LocalPort  string `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort string `json:"remote_port"`
	SshHost    string `json:"ssh_host"`
	SshPort    string `json:"ssh_port"`
	SshUser    string `json:"ssh_user"`
	SshKey     string `json:"ssh_key"`
	SshPass    string `json:"ssh_pass"`
	CreatedAt  int64  `json:"created_at"`
}

const configFile = "config.json"

func loadConfig() error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	tunnelMux.Lock()
	defer tunnelMux.Unlock()

	for _, t := range cfg.Tunnels {
		createdAt := t.CreatedAt
		if createdAt == 0 {
			createdAt = time.Now().Unix()
		}
		tunnels[t.Name] = &Tunnel{
			ID:         t.Name,
			Name:       t.Name,
			Type:       t.Type,
			LocalPort:  t.LocalPort,
			RemoteHost: t.RemoteHost,
			RemotePort: t.RemotePort,
			SshHost:    t.SshHost,
			SshPort:    t.SshPort,
			SshUser:    t.SshUser,
			SshKey:     t.SshKey,
			SshPass:    t.SshPass,
			Status:     "stopped",
			CreatedAt:  createdAt,
		}
	}

	return nil
}

func saveConfig() error {
	tunnelMux.RLock()
	defer tunnelMux.RUnlock()

	cfg := Config{
		Tunnels: make([]TunnelConfig, 0, len(tunnels)),
	}

	for _, t := range tunnels {
		cfg.Tunnels = append(cfg.Tunnels, TunnelConfig{
			Name:       t.Name,
			Type:       t.Type,
			LocalPort:  t.LocalPort,
			RemoteHost: t.RemoteHost,
			RemotePort: t.RemotePort,
			SshHost:    t.SshHost,
			SshPort:    t.SshPort,
			SshUser:    t.SshUser,
			SshKey:     t.SshKey,
			SshPass:    t.SshPass,
			CreatedAt:  t.CreatedAt,
		})
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
