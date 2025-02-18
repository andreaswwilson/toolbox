package mr

import (
	"context"
	"time"
	"toolbox/internal/config"

	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func NewAutoApplyCmd(ctx context.Context) *cobra.Command {
	var search string

	cmd := &cobra.Command{
		Use:   "auto-apply",
		Short: "Auto apply merge requests",
		Run: func(cmd *cobra.Command, args []string) {
			configData := ctx.Value(config.ConfigKey).(*config.Config)
			configData.Logger.Debug("Fetching all merge requests")
			client := configData.Client
			user, _, err := client.Users.CurrentUser()
			if err != nil {
				configData.Logger.Error("Failed to get current user", "error", err)
				return
			}

			// This only fetches the first page of merge requests. If you have more that 20 MR these won't be fetched without pagination.
			// I usually don't have 20+ MRs so I didn't bother to implement pagination.
			mergeRequests, _, err := client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
				AssigneeID: gitlab.AssigneeID(user.ID),
				Search:     &search,
				State:      gitlab.Ptr("opened"),
			})
			if err != nil {
				configData.Logger.Error("Failed to list merge requests", "error", err)
				return
			}
			configData.Logger.Info("Found merge requests", "count", len(mergeRequests))
			for _, mergeRequest := range mergeRequests {
				project, _, err := client.Projects.GetProject(mergeRequest.ProjectID, &gitlab.GetProjectOptions{})
				if err != nil {
					configData.Logger.Error("Failed to get project", "error", err)
					return
				}
				configData.Logger.Info("Checking merge request", "project", project.PathWithNamespace, "title", mergeRequest.Title, "status", mergeRequest.DetailedMergeStatus)
				if mergeRequest.DetailedMergeStatus == "mergeable" {
					configData.Logger.Info("Merging merge request", "project", project.PathWithNamespace, "title", mergeRequest.Title)
					_, _, err = client.MergeRequests.AcceptMergeRequest(mergeRequest.ProjectID, mergeRequest.IID, &gitlab.AcceptMergeRequestOptions{})
					if err != nil {
						configData.Logger.Error("Failed to merge merge request", "error", err)
						return
					}
					time.Sleep(5 * time.Second)
					mainPipeline, _, err := client.Pipelines.GetLatestPipeline(mergeRequest.ProjectID, &gitlab.GetLatestPipelineOptions{Ref: gitlab.Ptr("main")})
					if err != nil {
						configData.Logger.Error("Failed to get main pipeline", "error", err)
						return
					}
					for mainPipeline.Status != "success" {
						switch mainPipeline.Status {
						case "created", "waiting_for_resource", "preparing", "pending", "manual", "running":
							// This are ok statuses. Do nothing.
						default:
							configData.Logger.Error("Main pipeline failed", "status", mainPipeline.Status)
							return
						}
						jobs, _, err := client.Jobs.ListPipelineJobs(mergeRequest.ProjectID, mainPipeline.ID, &gitlab.ListJobsOptions{})
						if err != nil {
							configData.Logger.Error("Failed to list pipeline jobs", "error", err)
							return
						}
						for _, job := range jobs {
							switch job.Status {
							case "failed":
								configData.Logger.Error("Job failed", "name", job.Name)
								return
							case "manual":
								configData.Logger.Info("Playing job", "project", project.PathWithNamespace, "name", job.Name)
								_, _, err = client.Jobs.PlayJob(mergeRequest.ProjectID, job.ID, &gitlab.PlayJobOptions{})
								if err != nil {
									configData.Logger.Error("Failed to play job", "error", err)
									return
								}
							}
						}
						time.Sleep(time.Second * 5)
						mainPipeline, _, err = client.Pipelines.GetLatestPipeline(mergeRequest.ProjectID, &gitlab.GetLatestPipelineOptions{Ref: gitlab.Ptr("main")})
						if err != nil {
							configData.Logger.Error("Failed to get main pipeline", "error", err)
							return
						}
					}
				}
			}
		},
	}
	cmd.Flags().StringVarP(&search, "search", "s", "", "Search for merge requests")
	_ = cmd.MarkFlagRequired("search")

	return cmd
}
