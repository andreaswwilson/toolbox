package main

import (
	"context"
	"io"
	"os"
	"strings"
	"toolbox/cmd/gitlab/mr"
	"toolbox/cmd/gitlab/todo"
	"toolbox/internal/config"

	"log/slog"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var configData config.Config
var debug bool

var rootCmd = &cobra.Command{
	Use: "gl",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	ctx := context.Background()
	rootCmd.PersistentFlags().StringVarP(&configData.Token, "token", "t", "", "GitLab API token")
	rootCmd.PersistentFlags().StringVarP(&configData.BaseURL, "url", "u", "https://gitlab.skead.no", "GitLab instance URL")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	cobra.OnInitialize(initLogger, initClient)
	ctx = config.WithConfig(ctx, &configData)
	rootCmd.AddCommand(todo.NewTodoCmd(ctx))
	rootCmd.AddCommand(mr.NewMRCmd(ctx))
}

func initLogger() {
	var logger *slog.Logger
	if debug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		logger.Debug("Debug logging enabled")
	} else {
		// Set a logger that discards logs
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	configData.Logger = logger
}

func initClient() {
	if configData.Token == "" {
		service := "glab:" + strings.TrimPrefix(configData.BaseURL, "https://")
		if debug {
			configData.Logger.Debug("No token given. Looking for GitLab API token in Keychain", "service", service)
		}

		token, err := keyring.Get(service, "")
		if err != nil || token == "" {
			rootCmd.PrintErrln("GitLab API token is required. Please provide it using the --token flag")
			os.Exit(1)
		}
		configData.Token = string(token)
		if debug {
			configData.Logger.Debug("GitLab API token found in Keychain")
		}
	}

	var err error
	configData.Client, err = gitlab.NewClient(configData.Token, gitlab.WithBaseURL(configData.BaseURL))
	if err != nil {
		rootCmd.PrintErrln("Failed to create GitLab client:", err)
		os.Exit(1)
	}
	if debug {
		configData.Logger.Debug("GitLab client created successfully")
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if debug {
			configData.Logger.Error("Command execution failed", "error", err)
		}
		os.Exit(1)
	}
}

func main() {
	Execute()
}
