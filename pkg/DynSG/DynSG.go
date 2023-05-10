package DynSG

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rs/zerolog/log"
)

type SG struct {
	ctx    context.Context
	client *ec2.Client
	id     *string
	rules  []types.SecurityGroupRule
}

func New(id string) (sg *SG, err error) {
	sg.ctx = context.TODO()
	sg.id = &id

	// Setting up new Client
	awsConfig, err := config.LoadDefaultConfig(
		sg.ctx,
		config.WithSharedCredentialsFiles([]string{config.DefaultSharedCredentialsFilename()}),
		config.WithSharedConfigFiles([]string{config.DefaultSharedConfigFilename()}),
		config.WithSharedConfigProfile("DynSecurityGroup"),
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not initialize AWS Config")
		return
	}
	sg.client = ec2.NewFromConfig(awsConfig)

	filterType := "group-id"
	rules, err := sg.client.DescribeSecurityGroupRules(
		sg.ctx,
		&ec2.DescribeSecurityGroupRulesInput{
			Filters: []types.Filter{
				{
					Name:   &filterType,
					Values: []string{*sg.id},
				},
			},
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not retrieve rules for security group " + *sg.id)
		return
	}
	sg.rules = rules.SecurityGroupRules

	return
}

func (sg *SG) Update(ip string) (err error) {
	ipCidr := ip + "/32"
	request := make([]types.SecurityGroupRuleUpdate, 0, len(sg.rules))
	for _, rule := range sg.rules {
		request = append(
			request,
			types.SecurityGroupRuleUpdate{
				SecurityGroupRuleId: rule.SecurityGroupRuleId,
				SecurityGroupRule: &types.SecurityGroupRuleRequest{
					CidrIpv4:    &ipCidr,
					Description: rule.Description,
					FromPort:    rule.FromPort,
					ToPort:      rule.ToPort,
					IpProtocol:  rule.IpProtocol,
				},
			},
		)
	}
	_, err = sg.client.ModifySecurityGroupRules(
		sg.ctx,
		&ec2.ModifySecurityGroupRulesInput{
			GroupId:            sg.id,
			SecurityGroupRules: request,
		},
	)
	return
}
