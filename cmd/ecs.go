package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var ecsCluster string

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "List ECS services",
	Example: `  awsq ecs --cluster prod-cluster
  awsq ecs --cluster default -o json`,
	RunE: runECS,
}

func init() {
	ecsCmd.Flags().StringVarP(&ecsCluster, "cluster", "c", "default", "ECS cluster name")
	rootCmd.AddCommand(ecsCmd)
}

func runECS(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := ecs.NewFromConfig(cfg)

	// Collect all service ARNs across pages
	var allServiceArns []string
	listPaginator := ecs.NewListServicesPaginator(client, &ecs.ListServicesInput{
		Cluster: &ecsCluster,
	})
	for listPaginator.HasMorePages() {
		page, err := listPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list services: %w", err)
		}
		allServiceArns = append(allServiceArns, page.ServiceArns...)
	}

	if len(allServiceArns) == 0 {
		fmt.Printf("No services found in cluster '%s'\n", ecsCluster)
		return nil
	}

	// DescribeServices accepts max 10 at a time
	headers := []string{"SERVICE", "STATUS", "DESIRED", "RUNNING", "PENDING", "TASK_DEF"}
	var rows [][]string

	for i := 0; i < len(allServiceArns); i += 10 {
		end := i + 10
		if end > len(allServiceArns) {
			end = len(allServiceArns)
		}
		batch := allServiceArns[i:end]

		descResp, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  &ecsCluster,
			Services: batch,
		})
		if err != nil {
			return fmt.Errorf("failed to describe services: %w", err)
		}

		for _, svc := range descResp.Services {
			taskDef := derefStr(svc.TaskDefinition)
			if parts := splitLast(taskDef, "/"); parts != "" {
				taskDef = parts
			}

			rows = append(rows, []string{
				derefStr(svc.ServiceName),
				derefStr(svc.Status),
				fmt.Sprintf("%d", svc.DesiredCount),
				fmt.Sprintf("%d", svc.RunningCount),
				fmt.Sprintf("%d", svc.PendingCount),
				taskDef,
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d services in cluster '%s')\n", len(rows), ecsCluster)
	return nil
}

func splitLast(s string, sep string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if string(s[i]) == sep {
			return s[i+1:]
		}
	}
	return s
}
