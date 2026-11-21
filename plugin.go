package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	plugintypes "github.com/zerelio/plugin-sdk/plugin"
	sdk "github.com/zerelio/plugin-sdk"
)

type SlackPlugin struct {
	sdk.BasePlugin
}

type SlackConfig struct {
	BotToken string `json:"bot_token"`
}

type slackChannelsResponse struct {
	OK       bool            `json:"ok"`
	Channels []slackChannel  `json:"channels"`
	Error    string          `json:"error,omitempty"`
}

type slackChannel struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Topic          struct {
		Value string `json:"value"`
	} `json:"topic"`
	Purpose        struct {
		Value string `json:"value"`
	} `json:"purpose"`
	NumMembers     int    `json:"num_members"`
	IsArchived     bool   `json:"is_archived"`
	IsChannel      bool   `json:"is_channel"`
}

type slackUsersResponse struct {
	OK    bool          `json:"ok"`
	Members []slackUser `json:"members"`
	Error string        `json:"error,omitempty"`
}

type slackUser struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	RealName string `json:"real_name"`
	Profile struct {
		Email string `json:"email"`
	} `json:"profile"`
	Deleted bool `json:"deleted"`
	IsBot   bool `json:"is_bot"`
}

func (p *SlackPlugin) Info(ctx context.Context) (*plugintypes.PluginInfo, error) {
	return &plugintypes.PluginInfo{
		Name:        "slack",
		Version:     "1.0.0",
		Description: "Discovers Slack channels and users via the Slack Web API",
		Capabilities: []string{"communication", "collaboration"},
		CredentialTypes: []string{"bot_token"},
		ConfigSchema: `{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"required": ["bot_token"],
			"properties": {
				"bot_token": {"type": "string", "description": "Slack bot token (xoxb- prefix)"}
			}
		}`,
	}, nil
}

func (p *SlackPlugin) ValidateConfig(ctx context.Context, config map[string]string) (*plugintypes.ValidationResult, error) {
	var cfg SlackConfig
	if err := sdk.ParseConfig(config, &cfg); err != nil {
		return &plugintypes.ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("failed to parse config: %v", err)},
		}, nil
	}
	var errors []string
	if cfg.BotToken == "" {
		errors = append(errors, "bot_token is required")
	}
	if len(errors) > 0 {
		return &plugintypes.ValidationResult{Valid: false, Errors: errors}, nil
	}
	return &plugintypes.ValidationResult{Valid: true}, nil
}

func (p *SlackPlugin) Discover(ctx context.Context, req *plugintypes.DiscoverRequest) (<-chan plugintypes.DiscoveryEvent, error) {
	ch := make(chan plugintypes.DiscoveryEvent, 100)
	go func() {
		defer close(ch)

		var cfg SlackConfig
		if err := sdk.ParseConfig(req.Config, &cfg); err != nil {
			ch <- sdk.SendError("INVALID_CONFIG", fmt.Sprintf("failed to parse config: %v", err), false)
			return
		}
		if cfg.BotToken == "" {
			ch <- sdk.SendError("INVALID_CONFIG", "bot_token is required", false)
			return
		}

		client := &http.Client{Timeout: 30 * time.Second}
		ch <- sdk.SendProgress("Connecting to Slack", 5, 0, 0)

		// Discover channels
		ch <- sdk.SendProgress("Fetching channels", 25, 0, 0)
		channels, err := fetchSlackChannels(client, cfg.BotToken)
		if err != nil {
			ch <- sdk.SendError("API_ERROR", fmt.Sprintf("failed to fetch channels: %v", err), true)
		} else {
			for _, c := range channels {
				if c.IsArchived {
					continue
				}
				res := sdk.NewResource("slack", "slack.channel", "", "", c.ID)
				res.Properties = map[string]string{
					"name":        c.Name,
					"topic":       c.Topic.Value,
					"purpose":     c.Purpose.Value,
					"num_members": fmt.Sprintf("%d", c.NumMembers),
				}
				res.RawData = sdk.MarshalRawData(c)
				sdk.EmitResource(ch, res)
			}
		}

		// Discover users
		ch <- sdk.SendProgress("Fetching users", 50, int32(len(channels)), 0)
		users, err := fetchSlackUsers(client, cfg.BotToken)
		if err != nil {
			ch <- sdk.SendError("API_ERROR", fmt.Sprintf("failed to fetch users: %v", err), true)
		} else {
			for _, u := range users {
				if u.Deleted || u.IsBot {
					continue
				}
				res := sdk.NewResource("slack", "slack.user", "", "", u.ID)
				res.Properties = map[string]string{
					"name":      u.Name,
					"real_name": u.RealName,
					"email":     u.Profile.Email,
				}
				res.RawData = sdk.MarshalRawData(u)
				sdk.EmitResource(ch, res)
			}
		}

		total := int32(len(channels) + len(users))
		ch <- sdk.SendProgress("Slack discovery complete", 100, total, 0)
	}()
	return ch, nil
}

func slackAuthHeader(token string) string {
	return "Bearer " + token
}

func fetchSlackChannels(client *http.Client, botToken string) ([]slackChannel, error) {
	req, err := http.NewRequest("GET", "https://slack.com/api/conversations.list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", slackAuthHeader(botToken))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var result slackChannelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("Slack API error: %s", result.Error)
	}
	return result.Channels, nil
}

func fetchSlackUsers(client *http.Client, botToken string) ([]slackUser, error) {
	req, err := http.NewRequest("GET", "https://slack.com/api/users.list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", slackAuthHeader(botToken))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var result slackUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("Slack API error: %s", result.Error)
	}
	return result.Members, nil
}
