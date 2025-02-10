package todo

import (
	"context"

	"github.com/spf13/cobra"
)

func NewTodoCmd(ctx context.Context) *cobra.Command {
	var todoCmd = &cobra.Command{
		Use: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	todoCmd.AddCommand(NewClearTodoCmd(ctx))
	return todoCmd
}
