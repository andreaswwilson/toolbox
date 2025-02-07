package todo

import (
	"toolbox/internal/config"

	"github.com/spf13/cobra"
)

func NewTodoCmd(configData *config.Config) *cobra.Command {
	var todoCmd = &cobra.Command{
		Use: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	todoCmd.AddCommand(NewClearTodoCmd(configData))
	return todoCmd
}
