package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// AuthProfile represents a saved authentication profile.
type AuthProfile struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	Username         string `json:"username"`
	Password         string `json:"password,omitempty"`
	UsernameSelector string `json:"usernameSelector,omitempty"`
	PasswordSelector string `json:"passwordSelector,omitempty"`
	SubmitSelector   string `json:"submitSelector,omitempty"`
}

// AuthVault manages a local credential store.
type AuthVault struct {
	mu       sync.Mutex
	filepath string
	profiles map[string]AuthProfile
}

// NewAuthVault creates or loads the auth vault from ~/.ko-browser/auth.json.
func NewAuthVault() (*AuthVault, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	dir := filepath.Join(home, ".ko-browser")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	fp := filepath.Join(dir, "auth.json")

	v := &AuthVault{
		filepath: fp,
		profiles: make(map[string]AuthProfile),
	}

	if data, err := os.ReadFile(fp); err == nil {
		_ = json.Unmarshal(data, &v.profiles)
	}
	return v, nil
}

// Save stores a profile in the vault.
func (v *AuthVault) Save(p AuthProfile) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.profiles[p.Name] = p
	return v.flush()
}

// Get returns a profile by name.
func (v *AuthVault) Get(name string) (AuthProfile, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	p, ok := v.profiles[name]
	return p, ok
}

// Delete removes a profile by name.
func (v *AuthVault) Delete(name string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, ok := v.profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(v.profiles, name)
	return v.flush()
}

// List returns all profile names sorted.
func (v *AuthVault) List() []AuthProfile {
	v.mu.Lock()
	defer v.mu.Unlock()
	var result []AuthProfile
	for _, p := range v.profiles {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func (v *AuthVault) flush() error {
	data, err := json.MarshalIndent(v.profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(v.filepath, data, 0600)
}
