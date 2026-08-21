package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var rdsFilter string

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "List RDS database instances",
	Example: `  awsq rds
  awsq rds --filter engine=postgres
  awsq rds --filter status=available,engine=mysql
  awsq rds -r us-west-2 -o json`,
	RunE: runRDS,
}

func init() {
	rdsCmd.Flags().StringVarP(&rdsFilter, "filter", "f", "", "Filters: engine=postgres,status=available,multi_az=yes")
	rootCmd.AddCommand(rdsCmd)
}

func runRDS(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := rds.NewFromConfig(cfg)

	// Parse filters
	filters := parseGenericFilters(rdsFilter)

	headers := []string{"ID", "ENGINE", "VERSION", "CLASS", "STATUS", "MULTI_AZ", "STORAGE_GB", "ENDPOINT"}
	var rows [][]string

	paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe DB instances: %w", err)
		}

		for _, db := range page.DBInstances {
			multiAZ := "No"
			if db.MultiAZ != nil && *db.MultiAZ {
				multiAZ = "Yes"
			}

			engine := derefStr(db.Engine)
			status := derefStr(db.DBInstanceStatus)

			// Apply client-side filters
			if v, ok := filters["engine"]; ok && !strings.EqualFold(engine, v) {
				continue
			}
			if v, ok := filters["status"]; ok && !strings.EqualFold(status, v) {
				continue
			}
			if v, ok := filters["multi_az"]; ok && !strings.EqualFold(multiAZ, v) {
				continue
			}
			if v, ok := filters["class"]; ok && !strings.EqualFold(derefStr(db.DBInstanceClass), v) {
				continue
			}

			endpoint := "-"
			if db.Endpoint != nil && db.Endpoint.Address != nil {
				endpoint = fmt.Sprintf("%s:%d", *db.Endpoint.Address, *db.Endpoint.Port)
			}

			version := derefStr(db.EngineVersion)

			storage := "-"
			if db.AllocatedStorage != nil {
				storage = fmt.Sprintf("%d", *db.AllocatedStorage)
			}

			rows = append(rows, []string{
				derefStr(db.DBInstanceIdentifier),
				engine,
				version,
				derefStr(db.DBInstanceClass),
				status,
				multiAZ,
				storage,
				endpoint,
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d databases)\n", len(rows))
	return nil
}
