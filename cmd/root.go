package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"
)

var (
	region  string
	output  string
	profile string
)

var rootCmd = &cobra.Command{
	Use:   "awsq",
	Short: "awsq — query AWS resources in human-readable format",
	Long: `awsq is a fast, zero-setup CLI to query AWS resources without
complex jq chains. Think 'kubectl get' but for all AWS services.`,
	Version:           Version,
	PersistentPreRunE: resolveRegion,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "", "AWS region (defaults to AWS config/environment)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format: table, json, csv")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "AWS profile from ~/.aws/credentials")
	rootCmd.SetVersionTemplate("awsq {{.Version}}\n")
}

// resolveRegion determines the AWS region from flag, env, or AWS config.
// If none found, it returns an error asking the user to specify one.
func resolveRegion(cmd *cobra.Command, args []string) error {
	// If user explicitly passed --region, use it
	if region != "" {
		return nil
	}

	// Check AWS_DEFAULT_REGION and AWS_REGION env vars
	if r := os.Getenv("AWS_DEFAULT_REGION"); r != "" {
		region = r
		return nil
	}
	if r := os.Getenv("AWS_REGION"); r != "" {
		region = r
		return nil
	}

	// Try loading from ~/.aws/config (respects --profile and AWS_PROFILE)
	var opts []func(*config.LoadOptions) error
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err == nil && cfg.Region != "" {
		region = cfg.Region
		return nil
	}

	return fmt.Errorf("no AWS region found. Set one using:\n  --region / -r flag\n  AWS_DEFAULT_REGION environment variable\n  ~/.aws/config default region")
}

// GetAWSConfigOpts returns the config options for AWS SDK loading.
func GetAWSConfigOpts() []func(*config.LoadOptions) error {
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(region))
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	return opts
}

// derefStr safely dereferences a string pointer
func derefStr(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

// derefInt32 safely dereferences an int32 pointer
func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

// parseGenericFilters parses a "key=value,key2=value2" string into a map.
func parseGenericFilters(filterStr string) map[string]string {
	filters := make(map[string]string)
	if filterStr == "" {
		return filters
	}
	pairs := strings.Split(filterStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		}
	}
	return filters
}
