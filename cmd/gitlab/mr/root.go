package mr

import (
	"context"

	"github.com/spf13/cobra"
)

func NewMRCmd(ctx context.Context) *cobra.Command {
	mrCmd := &cobra.Command{
		Use: "mr",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	mrCmd.AddCommand(NewAutoApplyCmd(ctx))
	return mrCmd
}
