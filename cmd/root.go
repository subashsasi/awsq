package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	region string
	output string
)

var rootCmd = &cobra.Command{
	Use:   "awsq",
	Short: "awsq — query AWS resources in human-readable format",
	Long: `awsq is a fast, zero-setup CLI to query AWS resources without
complex jq chains. Think 'kubectl get' but for all AWS services.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "us-east-1", "AWS region")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format: table, json, csv")
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
