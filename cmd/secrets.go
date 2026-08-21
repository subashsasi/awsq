package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var secretsFilter string

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "List Secrets Manager secrets",
	Example: `  awsq secrets
  awsq secrets --filter prefix=prod/
  awsq secrets --filter rotation=enabled
  awsq secrets -r us-west-2 -o json`,
	RunE: runSecrets,
}

func init() {
	secretsCmd.Flags().StringVarP(&secretsFilter, "filter", "f", "", "Filters: prefix=prod/,rotation=enabled")
	rootCmd.AddCommand(secretsCmd)
}

func runSecrets(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Parse filters
	filters := parseGenericFilters(secretsFilter)

	headers := []string{"NAME", "LAST_ACCESSED", "LAST_ROTATED", "ROTATION", "DESCRIPTION"}
	var rows [][]string

	paginator := secretsmanager.NewListSecretsPaginator(client, &secretsmanager.ListSecretsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, secret := range page.SecretList {
			name := derefStr(secret.Name)

			rotation := "Disabled"
			if secret.RotationEnabled != nil && *secret.RotationEnabled {
				rotation = "Enabled"
			}

			// Apply client-side filters
			if v, ok := filters["prefix"]; ok && !strings.HasPrefix(name, v) {
				continue
			}
			if v, ok := filters["rotation"]; ok && !strings.EqualFold(rotation, v) {
				continue
			}

			lastAccessed := "-"
			if secret.LastAccessedDate != nil {
				lastAccessed = secret.LastAccessedDate.Format("2006-01-02")
			}

			lastRotated := "-"
			if secret.LastRotatedDate != nil {
				lastRotated = secret.LastRotatedDate.Format("2006-01-02")
			}

			desc := derefStr(secret.Description)
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}

			rows = append(rows, []string{
				name,
				lastAccessed,
				lastRotated,
				rotation,
				desc,
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d secrets)\n", len(rows))
	return nil
}
