package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PluginService struct {
	repo PluginRepository
}

func NewService(repo PluginRepository) *PluginService {
	return &PluginService{repo: repo}
}

func (s *PluginService) Create(name string, visible bool) (*Plugin, error) {
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	plug := &Plugin{ID: fmt.Sprintf("p-%d", time.Now().UnixNano()), Name: name, Visible: visible}
	if err := s.repo.Create(plug); err != nil {
		return nil, err
	}
	return plug, nil
}

func (s *PluginService) List() ([]*Plugin, error) {
	return s.repo.List()
}

func (s *PluginService) ListVisible() ([]*Plugin, error) {
	return s.repo.ListVisible()
}

func (s *PluginService) FindByID(id string) (*Plugin, error) {
	return s.repo.FindByID(id)
}

func (s *PluginService) CreateCustomCard(pluginID, name, data string) (*CustomCard, error) {
	if pluginID == "" || name == "" {
		return nil, fmt.Errorf("plugin_id and name required")
	}
	_, err := s.repo.FindByID(pluginID)
	if err != nil {
		return nil, err
	}
	card := &CustomCard{ID: fmt.Sprintf("cc-%d", time.Now().UnixNano()), PluginID: pluginID, Name: name, Data: data}
	if err := s.repo.CreateCustomCard(card); err != nil {
		return nil, err
	}
	return card, nil
}

func (s *PluginService) ListCustomCards(pluginID string) ([]*CustomCard, error) {
	if pluginID == "" {
		return nil, fmt.Errorf("plugin_id required")
	}
	return s.repo.ListCustomCards(pluginID)
}

func (s *PluginService) FindCustomCardByID(id string) (*CustomCard, error) {
	if id == "" {
		return nil, fmt.Errorf("card_id required")
	}
	return s.repo.FindCustomCardByID(id)
}

func (s *PluginService) SetUserPluginPermission(pluginID, userID string, allowed bool) (*UserPluginPermission, error) {
	if pluginID == "" || userID == "" {
		return nil, fmt.Errorf("plugin_id and user_id required")
	}
	_, err := s.repo.FindByID(pluginID)
	if err != nil {
		return nil, err
	}
	perm := &UserPluginPermission{ID: fmt.Sprintf("upp-%d", time.Now().UnixNano()), PluginID: pluginID, UserID: userID, Allowed: allowed}
	if err := s.repo.CreatePermission(perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *PluginService) GetUserPluginPermission(pluginID, userID string) (*UserPluginPermission, error) {
	if pluginID == "" || userID == "" {
		return nil, fmt.Errorf("plugin_id and user_id required")
	}
	return s.repo.GetPermission(pluginID, userID)
}

func (s *PluginService) HasPluginAccess(userID, pluginID string) (bool, error) {
	if userID == "" || pluginID == "" {
		return false, fmt.Errorf("user_id and plugin_id required")
	}
	perm, err := s.repo.GetPermission(pluginID, userID)
	if err != nil {
		return false, err
	}
	return perm != nil && perm.Allowed, nil
}

func (s *PluginService) DeleteUserPluginPermission(pluginID, userID string) error {
	if pluginID == "" || userID == "" {
		return fmt.Errorf("plugin_id and user_id required")
	}
	return s.repo.DeletePermission(pluginID, userID)
}

type SyncPluginRepoPayload struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Visible bool           `json:"visible"`
	RepoURL string         `json:"repo_url"`
	Cards   []CustomCard   `json:"cards"`
}

type remotePluginDocument struct {
	Plugins []SyncPluginRepoPayload `json:"plugins"`
}

func (s *PluginService) SyncRepository(repoURL string) ([]*Plugin, error) {
	if repoURL == "" {
		return nil, fmt.Errorf("repo URL required")
	}

	resp, err := http.Get(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed fetch repo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plugin repo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read repo body: %w", err)
	}

	var doc remotePluginDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		// Try direct plugin object array fallback.
		var arr []SyncPluginRepoPayload
		if err2 := json.Unmarshal(body, &arr); err2 != nil {
			return nil, fmt.Errorf("failed parse plugin repo JSON: %v", err)
		}
		doc.Plugins = arr
	}

	var synced []*Plugin
	for _, payload := range doc.Plugins {
		if payload.Name == "" {
			continue
		}
		pluginObj, err := s.repo.FindByID(payload.ID)
		if err != nil {
			// create new plugin if not exists.
			if payload.ID == "" {
				payload.ID = fmt.Sprintf("p-%d", time.Now().UnixNano())
			}
			pluginObj = &Plugin{ID: payload.ID, Name: payload.Name, Visible: payload.Visible, RepoURL: payload.RepoURL}
			if err = s.repo.Create(pluginObj); err != nil {
				return nil, err
			}
		} else {
			pluginObj.Name = payload.Name
			pluginObj.Visible = payload.Visible
			pluginObj.RepoURL = payload.RepoURL
			if err = s.repo.Update(pluginObj); err != nil {
				return nil, err
			}
		}

		for _, c := range payload.Cards {
			if c.ID == "" || c.PluginID == "" {
				c.PluginID = pluginObj.ID
			}
			if c.Name == "" {
				continue
			}
			if err = s.repo.UpsertCustomCard(&CustomCard{ID: c.ID, PluginID: pluginObj.ID, Name: c.Name, Data: c.Data}); err != nil {
				return nil, err
			}
		}

		synced = append(synced, pluginObj)
	}

	return synced, nil
}
