package main

import (
	"context"
	"testing"
)

func TestSlackPlugin_ValidateConfig(t *testing.T) {
	p := &SlackPlugin{}
	r, _ := p.ValidateConfig(context.Background(), map[string]string{})
	if r.Valid { t.Fatal("expected invalid") }
	r, _ = p.ValidateConfig(context.Background(), map[string]string{"bot_token": "xoxb-test"})
	if !r.Valid { t.Fatal("expected valid") }
}

func TestSlackPlugin_Info(t *testing.T) {
	p := &SlackPlugin{}
	info, _ := p.Info(context.Background())
	if info.Name != "slack" { t.Fatalf("expected slack, got %s", info.Name) }
}
