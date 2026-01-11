package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type Manager struct {
	config *oauth2.Config
	token  *oauth2.Token
	mu     sync.RWMutex
}

func NewManager(clientSecretFile string) (*Manager, error) {
	data, err := os.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("read client secret: %w", err)
	}

	config, err := google.ConfigFromJSON(data, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse client secret: %w", err)
	}

	return &Manager{config: config}, nil
}

func (m *Manager) AuthURL() (string, string) {
	state := generateState()
	url := m.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, state
}

func (m *Manager) Exchange(ctx context.Context, code string) error {
	token, err := m.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("exchange code: %w", err)
	}

	m.mu.Lock()
	m.token = token
	m.mu.Unlock()

	return nil
}

func (m *Manager) Token() *oauth2.Token {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token
}

func (m *Manager) Client(ctx context.Context) *oauth2.Token {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token
}

func (m *Manager) Config() *oauth2.Config {
	return m.config
}

func (m *Manager) IsAuthenticated() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token != nil && m.token.Valid()
}

func (m *Manager) Logout() {
	m.mu.Lock()
	m.token = nil
	m.mu.Unlock()
}

func (m *Manager) SaveToken(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.token == nil {
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(m.token)
}

func (m *Manager) LoadToken(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return err
	}

	m.mu.Lock()
	m.token = &token
	m.mu.Unlock()

	return nil
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
