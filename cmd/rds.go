package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "List RDS database instances",
	Example: `  awsq rds
  awsq rds -r us-west-2
  awsq rds -o json`,
	RunE: runRDS,
}

func init() {
	rootCmd.AddCommand(rdsCmd)
}

func runRDS(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := rds.NewFromConfig(cfg)

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

			endpoint := "-"
			if db.Endpoint != nil && db.Endpoint.Address != nil {
				endpoint = fmt.Sprintf("%s:%d", *db.Endpoint.Address, *db.Endpoint.Port)
			}

			engine := derefStr(db.Engine)
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
				derefStr(db.DBInstanceStatus),
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
