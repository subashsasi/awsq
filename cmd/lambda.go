package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var lambdaCmd = &cobra.Command{
	Use:   "lambda",
	Short: "List Lambda functions",
	Example: `  awsq lambda
  awsq lambda -r us-west-2
  awsq lambda -o json`,
	RunE: runLambda,
}

func init() {
	rootCmd.AddCommand(lambdaCmd)
}

func runLambda(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := lambda.NewFromConfig(cfg)

	headers := []string{"NAME", "RUNTIME", "MEMORY_MB", "TIMEOUT_S", "LAST_MODIFIED", "DESCRIPTION"}
	var rows [][]string

	paginator := lambda.NewListFunctionsPaginator(client, &lambda.ListFunctionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list functions: %w", err)
		}

		for _, fn := range page.Functions {
			desc := derefStr(fn.Description)
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}

			rows = append(rows, []string{
				derefStr(fn.FunctionName),
				string(fn.Runtime),
				fmt.Sprintf("%d", derefInt32(fn.MemorySize)),
				fmt.Sprintf("%d", derefInt32(fn.Timeout)),
				derefStr(fn.LastModified),
				desc,
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d functions)\n", len(rows))
	return nil
}

