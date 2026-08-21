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

var ec2Filter string

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "List EC2 instances",
	Example: `  awsq ec2
  awsq ec2 --filter state=running
  awsq ec2 --filter state=running,type=t3.micro
  awsq ec2 --filter name=web-server
  awsq ec2 -r us-west-2 -o json`,
	RunE: runEC2,
}

func init() {
	ec2Cmd.Flags().StringVarP(&ec2Filter, "filter", "f", "", "Filters: state=running,type=t3.micro,name=web-server")
	rootCmd.AddCommand(ec2Cmd)
}

func runEC2(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, GetAWSConfigOpts()...)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	input := &ec2.DescribeInstancesInput{}

	if ec2Filter != "" {
		input.Filters = parseEC2Filters(ec2Filter)
	}

	headers := []string{"ID", "NAME", "TYPE", "STATE", "PRIVATE_IP", "PUBLIC_IP", "AZ"}
	var rows [][]string

	paginator := ec2.NewDescribeInstancesPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe instances: %w", err)
		}

		for _, reservation := range page.Reservations {
			for _, inst := range reservation.Instances {
				rows = append(rows, []string{
					derefStr(inst.InstanceId),
					getTagValue(inst.Tags, "Name"),
					string(inst.InstanceType),
					string(inst.State.Name),
					derefStr(inst.PrivateIpAddress),
					derefStr(inst.PublicIpAddress),
					getPlacementAZ(inst.Placement),
				})
			}
		}
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d instances)\n", len(rows))
	return nil
}

func parseEC2Filters(filterStr string) []types.Filter {
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
		case "state":
			awsKey = "instance-state-name"
		case "type":
			awsKey = "instance-type"
		case "name":
			awsKey = "tag:Name"
		case "vpc":
			awsKey = "vpc-id"
		case "az":
			awsKey = "availability-zone"
		default:
			awsKey = "tag:" + key
		}

		filters = append(filters, types.Filter{
			Name:   &awsKey,
			Values: []string{value},
		})
	}
	return filters
}

func getTagValue(tags []types.Tag, key string) string {
	for _, tag := range tags {
		if tag.Key != nil && *tag.Key == key {
			return derefStr(tag.Value)
		}
	}
	return "-"
}

func getPlacementAZ(placement *types.Placement) string {
	if placement == nil {
		return "-"
	}
	return derefStr(placement.AvailabilityZone)
}
