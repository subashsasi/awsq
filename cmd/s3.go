package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var s3Filter string

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "List S3 buckets",
	Example: `  awsq s3
  awsq s3 --filter prefix=prod
  awsq s3 --filter prefix=logs-
  awsq s3 -o json`,
	RunE: runS3,
}

func init() {
	s3Cmd.Flags().StringVarP(&s3Filter, "filter", "f", "", "Filters: prefix=prod")
	rootCmd.AddCommand(s3Cmd)
}

func runS3(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	resp, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to list buckets: %w", err)
	}

	// Parse prefix filter
	var prefix string
	if s3Filter != "" {
		for _, pair := range strings.Split(s3Filter, ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 && parts[0] == "prefix" {
				prefix = parts[1]
			}
		}
	}

	headers := []string{"NAME", "CREATED"}
	var rows [][]string

	for _, bucket := range resp.Buckets {
		name := derefStr(bucket.Name)

		// Apply prefix filter (client-side)
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}

		created := "-"
		if bucket.CreationDate != nil {
			created = bucket.CreationDate.Format("2006-01-02 15:04")
		}
		rows = append(rows, []string{
			name,
			created,
		})
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d buckets)\n", len(rows))
	return nil
}
