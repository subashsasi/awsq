package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var sgFilter string

var sgCmd = &cobra.Command{
	Use:   "sg",
	Short: "List Security Groups",
	Example: `  awsq sg
  awsq sg --filter vpc=vpc-123abc
  awsq sg -o json`,
	RunE: runSG,
}

func init() {
	sgCmd.Flags().StringVarP(&sgFilter, "filter", "f", "", "Filters: vpc=vpc-123abc,name=my-sg")
	rootCmd.AddCommand(sgCmd)
}

func runSG(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeSecurityGroupsInput{}
	if sgFilter != "" {
		input.Filters = parseSGFilters(sgFilter)
	}

	headers := []string{"ID", "NAME", "VPC", "INBOUND_RULES", "OUTBOUND_RULES", "DESCRIPTION"}
	var rows [][]string

	paginator := ec2.NewDescribeSecurityGroupsPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe security groups: %w", err)
		}

		for _, sg := range page.SecurityGroups {
			rows = append(rows, []string{
				derefStr(sg.GroupId),
				derefStr(sg.GroupName),
				derefStr(sg.VpcId),
				fmt.Sprintf("%d", len(sg.IpPermissions)),
				fmt.Sprintf("%d", len(sg.IpPermissionsEgress)),
				derefStr(sg.Description),
			})
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d security groups)\n", len(rows))
	return nil
}

func parseSGFilters(filterStr string) []types.Filter {
	var filters []types.Filter
	pairs := strings.Split(filterStr, ",")

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		var awsKey string
		switch key {
		case "vpc":
			awsKey = "vpc-id"
		case "name":
			awsKey = "group-name"
		case "id":
			awsKey = "group-id"
		default:
			awsKey = key
		}

		filters = append(filters, types.Filter{
			Name:   &awsKey,
			Values: []string{value},
		})
	}
	return filters
}
