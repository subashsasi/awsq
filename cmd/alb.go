package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var albCmd = &cobra.Command{
	Use:   "alb",
	Short: "List Application Load Balancers",
	Example: `  awsq alb
  awsq alb -r us-west-2
  awsq alb -o json`,
	RunE: runALB,
}

func init() {
	rootCmd.AddCommand(albCmd)
}

func runALB(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := elbv2.NewFromConfig(cfg)

	headers := []string{"NAME", "TYPE", "SCHEME", "STATE", "DNS_NAME", "VPC"}
	var rows [][]string

	paginator := elbv2.NewDescribeLoadBalancersPaginator(client, &elbv2.DescribeLoadBalancersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe load balancers: %w", err)
		}

		for _, lb := range page.LoadBalancers {
			state := "-"
			if lb.State != nil {
				state = string(lb.State.Code)
			}

			rows = append(rows, []string{
				derefStr(lb.LoadBalancerName),
				string(lb.Type),
				string(lb.Scheme),
				state,
				derefStr(lb.DNSName),
				derefStr(lb.VpcId),
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d load balancers)\n", len(rows))
	return nil
}
