package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var albFilter string

var albCmd = &cobra.Command{
	Use:   "alb",
	Short: "List Application Load Balancers",
	Example: `  awsq alb
  awsq alb --filter scheme=internet-facing
  awsq alb --filter type=application,scheme=internal
  awsq alb -r us-west-2 -o json`,
	RunE: runALB,
}

func init() {
	albCmd.Flags().StringVarP(&albFilter, "filter", "f", "", "Filters: scheme=internet-facing,type=application,vpc=vpc-123")
	rootCmd.AddCommand(albCmd)
}

func runALB(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := elbv2.NewFromConfig(cfg)

	// Parse filters
	filters := parseGenericFilters(albFilter)

	headers := []string{"NAME", "TYPE", "SCHEME", "STATE", "DNS_NAME", "VPC"}
	var rows [][]string

	paginator := elbv2.NewDescribeLoadBalancersPaginator(client, &elbv2.DescribeLoadBalancersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe load balancers: %w", err)
		}

		for _, lb := range page.LoadBalancers {
			lbType := string(lb.Type)
			scheme := string(lb.Scheme)
			vpcID := derefStr(lb.VpcId)

			// Apply client-side filters
			if v, ok := filters["type"]; ok && !strings.EqualFold(lbType, v) {
				continue
			}
			if v, ok := filters["scheme"]; ok && !strings.EqualFold(scheme, v) {
				continue
			}
			if v, ok := filters["vpc"]; ok && vpcID != v {
				continue
			}

			state := "-"
			if lb.State != nil {
				state = string(lb.State.Code)
			}

			rows = append(rows, []string{
				derefStr(lb.LoadBalancerName),
				lbType,
				scheme,
				state,
				derefStr(lb.DNSName),
				vpcID,
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d load balancers)\n", len(rows))
	return nil
}
