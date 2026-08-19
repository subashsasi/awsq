package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/subashsasi/awsq/pkg/formatter"
)

var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "List VPCs",
	Example: `  awsq vpc
  awsq vpc -r us-west-2
  awsq vpc -o json`,
	RunE: runVPC,
}

func init() {
	rootCmd.AddCommand(vpcCmd)
}

func runVPC(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := ec2.NewFromConfig(cfg)

	// Get all VPCs with pagination
	var allVpcs []types.Vpc
	vpcPaginator := ec2.NewDescribeVpcsPaginator(client, &ec2.DescribeVpcsInput{})
	for vpcPaginator.HasMorePages() {
		page, err := vpcPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe VPCs: %w", err)
		}
		allVpcs = append(allVpcs, page.Vpcs...)
	}

	// Get all subnets with pagination to count per VPC
	subnetCount := make(map[string]int)
	subnetPaginator := ec2.NewDescribeSubnetsPaginator(client, &ec2.DescribeSubnetsInput{})
	for subnetPaginator.HasMorePages() {
		page, err := subnetPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe subnets: %w", err)
		}
		for _, subnet := range page.Subnets {
			vpcID := derefStr(subnet.VpcId)
			subnetCount[vpcID]++
		}
	}

	headers := []string{"VPC_ID", "NAME", "CIDR", "STATE", "DEFAULT", "SUBNETS"}
	var rows [][]string

	for _, vpc := range allVpcs {
		vpcID := derefStr(vpc.VpcId)
		name := getVPCTagValue(vpc.Tags, "Name")

		isDefault := "No"
		if vpc.IsDefault != nil && *vpc.IsDefault {
			isDefault = "Yes"
		}

		rows = append(rows, []string{
			vpcID,
			name,
			derefStr(vpc.CidrBlock),
			string(vpc.State),
			isDefault,
			fmt.Sprintf("%d", subnetCount[vpcID]),
		})
	}

	formatter.Print(output, headers, rows)
	fmt.Printf("\n(%d VPCs)\n", len(rows))
	return nil
}

func getVPCTagValue(tags []types.Tag, key string) string {
	for _, tag := range tags {
		if tag.Key != nil && *tag.Key == key {
			return derefStr(tag.Value)
		}
	}
	return "-"
}
