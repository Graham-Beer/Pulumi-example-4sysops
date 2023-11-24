package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		config := config.New(ctx, "")

		// Stack values
		IngressIp := config.Require("IngressIp")
		VpcId := config.Require("VpcId")
		SubnetId := config.Require("SubnetId")

		// Getting latest Amazon Linux 2 image
		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			Filters: []ec2.GetAmiFilter{
				{
					Name:   "name",                                     // Use a simple string literal for the filter name
					Values: []string{"amzn2-ami-hvm-2.0.*-x86_64-gp2"}, // Pass a string slice directly for the filter values
				},
			},
			MostRecent: pulumi.BoolRef(true), // Pass a `bool` pointer directly without using pulumi.BoolPtr
			Owners:     []string{"amazon"},   // Again, pass a string slice directly without wrapping
		})
		if err != nil {
			return err
		}

		// Security group which allows SSH traffic
		sg, err := ec2.NewSecurityGroup(ctx, "ec2SecurityGroup", &ec2.SecurityGroupArgs{
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
					Protocol: pulumi.String("-1"),
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
				},
			},
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					CidrBlocks: pulumi.StringArray{
						pulumi.String(IngressIp),
					},
					Protocol: pulumi.String("tcp"),
					FromPort: pulumi.Int(22), // SSH port
					ToPort:   pulumi.Int(22),
				},
			},
			VpcId: pulumi.String(VpcId),
		})
		if err != nil {
			return err
		}

		// Create EC2 instance with Security group and image
		ec2Instance, err := ec2.NewInstance(ctx, "ec2Instance", &ec2.InstanceArgs{
			Ami:                 pulumi.String(ami.Id),
			InstanceType:        pulumi.String("t2.micro"),
			SubnetId:            pulumi.String(SubnetId),
			VpcSecurityGroupIds: pulumi.StringArray{sg.ID()},
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Pulumi-example"),
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("InstanceId", ec2Instance.ID())

		return nil
	})
}
