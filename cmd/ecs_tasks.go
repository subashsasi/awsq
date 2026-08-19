package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var (
	ecsTasksCluster string
	ecsTasksService string
)

var ecsTasksCmd = &cobra.Command{
	Use:   "ecs-tasks",
	Short: "List ECS tasks in a service",
	Example: `  awsq ecs-tasks --cluster prod --service my-service
  awsq ecs-tasks -c prod -s payment-service -o json`,
	RunE: runECSTasks,
}

func init() {
	ecsTasksCmd.Flags().StringVarP(&ecsTasksCluster, "cluster", "c", "default", "ECS cluster name")
	ecsTasksCmd.Flags().StringVarP(&ecsTasksService, "service", "s", "", "ECS service name (required)")
	ecsTasksCmd.MarkFlagRequired("service")
	rootCmd.AddCommand(ecsTasksCmd)
}

func runECSTasks(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := ecs.NewFromConfig(cfg)

	// Collect all task ARNs across pages
	var allTaskArns []string
	listPaginator := ecs.NewListTasksPaginator(client, &ecs.ListTasksInput{
		Cluster:     &ecsTasksCluster,
		ServiceName: &ecsTasksService,
	})
	for listPaginator.HasMorePages() {
		page, err := listPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}
		allTaskArns = append(allTaskArns, page.TaskArns...)
	}

	if len(allTaskArns) == 0 {
		fmt.Printf("No tasks found for service '%s' in cluster '%s'\n", ecsTasksService, ecsTasksCluster)
		return nil
	}

	// DescribeTasks accepts max 100 at a time
	headers := []string{"TASK_ID", "STATUS", "HEALTH", "LAUNCH_TYPE", "CPU", "MEMORY", "STARTED"}
	var rows [][]string

	for i := 0; i < len(allTaskArns); i += 100 {
		end := i + 100
		if end > len(allTaskArns) {
			end = len(allTaskArns)
		}
		batch := allTaskArns[i:end]

		descResp, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
			Cluster: &ecsTasksCluster,
			Tasks:   batch,
		})
		if err != nil {
			return fmt.Errorf("failed to describe tasks: %w", err)
		}

		for _, task := range descResp.Tasks {
			taskID := derefStr(task.TaskArn)
			if idx := strings.LastIndex(taskID, "/"); idx != -1 {
				taskID = taskID[idx+1:]
			}

			health := string(task.HealthStatus)
			if health == "" {
				health = "-"
			}

			started := "-"
			if task.StartedAt != nil {
				started = task.StartedAt.Format("2006-01-02 15:04")
			}

			launchType := string(task.LaunchType)
			if launchType == "" {
				launchType = "-"
			}

			rows = append(rows, []string{
				taskID,
				derefStr(task.LastStatus),
				health,
				launchType,
				derefStr(task.Cpu),
				derefStr(task.Memory),
				started,
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d tasks)\n", len(rows))
	return nil
}


