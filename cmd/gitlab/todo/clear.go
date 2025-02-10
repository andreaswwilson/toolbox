package todo

import (
	"context"
	"log/slog"
	"toolbox/internal/config"

	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func NewClearTodoCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all todos",
		Run: func(cmd *cobra.Command, args []string) {
			configData := ctx.Value(config.ConfigKey).(*config.Config)
			// Use the client from the config to get all todos
			if configData.Client == nil {
				configData.Logger.Error("GitLab client is not initialized")
				cmd.PrintErrln("GitLab client is not initialized")
				return
			}
			configData.Logger.Debug("Fetching all todos")
			todos, err := getAllTodos(ctx)
			if err != nil {
				configData.Logger.Error("Failed to list todos", "error", err)
				cmd.PrintErrln("Failed to list todos")
				return
			}
			for _, todo := range todos {
				configData.Logger.Debug("Processing todo", "todo", todo)
				mergeRequest, err := getMergeRequestForTodo(ctx, todo)
				if err != nil {
					configData.Logger.Error("Failed to get merge request for todo", "error", err)
					cmd.PrintErrln("Failed to get merge request for todo")
					continue
				}
				if shouldMarkTodoAsDone(todo, mergeRequest) {
					markTodoAsDone(ctx, cmd, todo)
				}
			}
		},
	}
}

func getAllTodos(ctx context.Context) ([]*gitlab.Todo, error) {
	configData := ctx.Value(config.ConfigKey).(*config.Config)
	configData.Logger.Debug("Get all todos")
	todos, _, err := configData.Client.Todos.ListTodos(&gitlab.ListTodosOptions{})
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func getMergeRequestForTodo(ctx context.Context, todo *gitlab.Todo) (*gitlab.MergeRequest, error) {
	configData := ctx.Value(config.ConfigKey).(*config.Config)
	configData.Logger.Debug("Fetching merge request for todo", "todo", todo)
	mergeRequests, _, err := configData.Client.MergeRequests.ListProjectMergeRequests(todo.Project.ID, &gitlab.ListProjectMergeRequestsOptions{
		SourceBranch: &todo.Target.SourceBranch,
		TargetBranch: &todo.Target.TargetBranch,
		AuthorID:     &todo.Target.Author.ID,
	})
	if err != nil {
		return nil, err
	}
	if len(mergeRequests) != 1 {
		return nil, nil
	}
	return mergeRequests[0], nil
}

func shouldMarkTodoAsDone(todo *gitlab.Todo, mergeRequest *gitlab.MergeRequest) bool {
	if todo == nil {
		slog.Debug("shouldMarkTodoAsDone", "message", "todo is nil")
		return false
	}
	if todo.ActionName != "review_requested" {
		slog.Debug("shouldMarkTodoAsDone", "message", "ActionName is not review_requested", "ActionName", todo.ActionName)
		return false
	}
	if todo.Target == nil {
		slog.Debug("shouldMarkTodoAsDone", "message", "Target is nil")
		return false
	}
	if todo.Target.MergedAt != nil {
		slog.Debug("shouldMarkTodoAsDone", "message", "Target is merged", "MergedAt", todo.Target.MergedAt)
		return true
	}
	if todo.Target.State == "closed" {
		slog.Debug("shouldMarkTodoAsDone", "message", "Target is closed", "State", todo.Target.State)
		return true
	}
	if mergeRequest != nil && mergeRequest.DetailedMergeStatus == "mergeable" {
		slog.Debug("shouldMarkTodoAsDone", "message", "MergeRequest is mergeable", "DetailedMergeStatus", mergeRequest.DetailedMergeStatus)
		return true
	}
	slog.Debug("shouldMarkTodoAsDone", "message", "None of the conditions met")
	return false
}

func markTodoAsDone(ctx context.Context, cmd *cobra.Command, todo *gitlab.Todo) {
	configData := ctx.Value(config.ConfigKey).(*config.Config)
	configData.Logger.Debug("Marking todo as done", "todo", todo)
	cmd.Println("Marking as done.", "Author:", todo.Target.Author.Name, "Source brance:", todo.Target.SourceBranch, "Project:", todo.Project.Name)
	_, err := configData.Client.Todos.MarkTodoAsDone(todo.ID)
	if err != nil {
		configData.Logger.Error("Failed to mark todo as done", "error", err)
	}
}
