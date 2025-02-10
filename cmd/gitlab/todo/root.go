package todo

import (
	"context"
	"toolbox/internal/config"

	"github.com/spf13/cobra"
)

func NewTodoCmd(configData *config.Config) *cobra.Command {
	ctx := context.WithValue(context.Background(), config.ConfigKey, configData)
	var todoCmd = &cobra.Command{
		Use: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	todoCmd.AddCommand(NewClearTodoCmd(ctx))
	return todoCmd
}
