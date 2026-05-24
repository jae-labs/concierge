package main

import (
	"log/slog"
	"os"

	"github.com/jae-labs/conCIerge/internal/config"
	ghclient "github.com/jae-labs/conCIerge/internal/github"
	slackhandler "github.com/jae-labs/conCIerge/internal/slack"
	slacklib "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	api := slacklib.New(
		cfg.SlackBotToken,
		slacklib.OptionAppLevelToken(cfg.SlackAppToken),
	)

	sm := socketmode.New(api)

	gh, err := ghclient.NewClient(
		cfg.GitHubAppID, cfg.GitHubAppInstallationID, cfg.GitHubAppPrivateKey,
		cfg.GitHubOwner, cfg.GitHubRepo,
	)
	if err != nil {
		slog.Error("failed to create github client", "error", err)
		os.Exit(1)
	}

	handler := slackhandler.NewHandler(
		api,
		sm,
		gh,
		cfg.SlackRequestsChannelID,
		cfg.SlackUserIDs,
		cfg.SlackManagerIDs,
		cfg.SlackAdminIDs,
		logger,
	)

	slog.Info("concierge starting in socket mode")
	handler.Run()
}
