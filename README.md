# Zerelio Slack Plugin

Discovers Slack workspace resources and streams them into the Zerelio knowledge graph.

## What It Discovers

| Resource type          | Description                             |
|------------------------|-----------------------------------------|
| `slack.workspace`      | Workspace metadata and settings         |
| `slack.channel.public` | Public channels                         |
| `slack.channel.private` | Private channels (requires scope)      |
| `slack.user`           | Workspace members                       |
| `slack.bot`            | Bot users and app integrations          |

## Configuration

| Environment variable | Required | Description                                              |
|----------------------|----------|----------------------------------------------------------|
| `SLACK_BOT_TOKEN`    | Yes      | Bot OAuth token (starts with `xoxb-`)                   |

### Required Bot Token Scopes

The bot token must have the following OAuth scopes:

- `channels:read` — list public channels
- `groups:read` — list private channels the bot is a member of
- `users:read` — list workspace members
- `team:read` — read workspace metadata

To create a bot token: go to [api.slack.com/apps](https://api.slack.com/apps), create an app, add the scopes above under **OAuth & Permissions**, and install it to your workspace.

## Adding to a Zerelio Scope

```json
{
  "scope_name": "My Slack Workspace",
  "plugin_name": "slack",
  "config": {
    "workspace": "my-company"
  },
  "schedule": "@every 1h"
}
```

## Part of Zerelio

This plugin is part of the [Zerelio](https://github.com/zerelio/knowledge) infrastructure knowledge graph platform.
