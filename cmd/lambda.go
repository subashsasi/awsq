package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var lambdaFilter string

var lambdaCmd = &cobra.Command{
	Use:   "lambda",
	Short: "List Lambda functions",
	Example: `  awsq lambda
  awsq lambda --filter runtime=python3.12
  awsq lambda --filter prefix=payment
  awsq lambda -r us-west-2 -o json`,
	RunE: runLambda,
}

func init() {
	lambdaCmd.Flags().StringVarP(&lambdaFilter, "filter", "f", "", "Filters: runtime=python3.12,prefix=payment")
	rootCmd.AddCommand(lambdaCmd)
}

func runLambda(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := lambda.NewFromConfig(cfg)

	// Parse filters
	filters := parseGenericFilters(lambdaFilter)

	headers := []string{"NAME", "RUNTIME", "MEMORY_MB", "TIMEOUT_S", "LAST_MODIFIED", "DESCRIPTION"}
	var rows [][]string

	paginator := lambda.NewListFunctionsPaginator(client, &lambda.ListFunctionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list functions: %w", err)
		}

		for _, fn := range page.Functions {
			name := derefStr(fn.FunctionName)
			runtime := string(fn.Runtime)

			// Apply client-side filters
			if v, ok := filters["prefix"]; ok && !strings.HasPrefix(name, v) {
				continue
			}
			if v, ok := filters["runtime"]; ok && !strings.EqualFold(runtime, v) {
				continue
			}

			desc := derefStr(fn.Description)
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}

			rows = append(rows, []string{
				name,
				runtime,
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
