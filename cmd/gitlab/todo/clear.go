package todo

import (
	"toolbox/internal/config"

	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func NewClearTodoCmd(configData *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all todos",
		Run: func(cmd *cobra.Command, args []string) {
			// Use the client from the config to get all todos
			if configData.Client == nil {
				configData.Logger.Error("GitLab client is not initialized")
				cmd.PrintErrln("GitLab client is not initialized")
				return
			}
			todos, _, err := configData.Client.Todos.ListTodos(&gitlab.ListTodosOptions{})
			if err != nil {
				configData.Logger.Error("Failed to list todos", "error", err)
				cmd.PrintErrln("Failed to list todos")
				return
			}
			for _, todo := range todos {
				if todo.ActionName != "review_requested" {
					continue
				}
				if todo.Target.MergedAt != nil {
					// mark todo as done if the merge request is already merged
					// configData.Logger.Debug("Marking todo as done", "author", todo.Target.Author.Name, "source_branch", todo.Target.SourceBranch, "project", todo.Project.Name)

					cmd.Println("Marking merged review request as done.", "Author:", todo.Target.Author.Name, "Source brance:", todo.Target.SourceBranch, "Project:", todo.Project.Name)
					_, err := configData.Client.Todos.MarkTodoAsDone(todo.ID)
					if err != nil {
						configData.Logger.Error("Failed to mark todo as done", "error", err)
					}
					continue
				}
				mergeRequests, _, err := configData.Client.MergeRequests.ListProjectMergeRequests(todo.Project.ID, &gitlab.ListProjectMergeRequestsOptions{
					SourceBranch: &todo.Target.SourceBranch,
					TargetBranch: &todo.Target.TargetBranch,
					AuthorID:     &todo.Target.Author.ID,
				})
				if err != nil {
					configData.Logger.Error("Failed to list merge requests", "error", err)
					cmd.PrintErrln("Failed to list merge requests")
					return
				}
				if len(mergeRequests) != 1 {
					configData.Logger.Debug("Merge request not found", "todo", todo, "expected to find 1 merge request, but found", len(mergeRequests))
					continue
				}
				mergeRequest := mergeRequests[0]
				if mergeRequest.DetailedMergeStatus == "mergeable" {
					configData.Logger.Debug("Marking todo as done", "todo", todo)
					_, err := configData.Client.Todos.MarkTodoAsDone(todo.ID)
					if err != nil {
						configData.Logger.Error("Failed to mark todo as done", "error", err)
						continue
					}
				}
			}
		},
	}
}
