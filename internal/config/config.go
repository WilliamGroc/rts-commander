package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"rtscommander/m/internal/remote"
)

// Config représente la configuration de l'application
type Config struct {
	Remotes    map[string]*remote.Control `json:"remotes"`
	ConfigPath string                     `json:"-"`
	mu         sync.RWMutex               `json:"-"`
}

// Load charge la configuration depuis un fichier JSON
func Load(path string) (*Config, error) {
	config := &Config{
		ConfigPath: path,
		Remotes:    make(map[string]*remote.Control),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Config file not found, creating new one: %s", path)
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	if err := json.Unmarshal(data, &config.Remotes); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	log.Printf("Loaded %d remote(s) from %s", len(config.Remotes), path)
	return config, nil
}

// Save sauvegarde la configuration dans le fichier JSON
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.Remotes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(c.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}

// AddRemote ajoute ou met à jour une télécommande
func (c *Config) AddRemote(name string, rc *remote.Control) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rc.Name = name
	c.Remotes[name] = rc

	return c.Save()
}

// ListRemotes retourne la liste des noms de télécommandes
func (c *Config) ListRemotes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.Remotes))
	for name := range c.Remotes {
		names = append(names, name)
	}
	return names
}

// GetRemote retourne une télécommande par son nom
func (c *Config) GetRemote(name string) (*remote.Control, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rc, exists := c.Remotes[name]
	return rc, exists
}
