package DynR53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/rs/zerolog/log"
)

type R53 struct {
	ctx       context.Context
	client    *route53.Client
	domainID  *string
	recordSet *types.ResourceRecordSet
}

func New(domainName, subDomainName string) (r *R53, err error) {
	r.ctx = context.TODO()

	// Setting up new Client
	awsConfig, err := config.LoadDefaultConfig(
		r.ctx,
		config.WithSharedCredentialsFiles([]string{config.DefaultSharedCredentialsFilename()}),
		config.WithSharedConfigFiles([]string{config.DefaultSharedConfigFilename()}),
		config.WithSharedConfigProfile("DynRoute53"),
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not initialize AWS Config")
		return
	}
	r.client = route53.NewFromConfig(awsConfig)

	// Finding hosted zone ID
	allZones, err := r.client.ListHostedZonesByName(
		r.ctx,
		&route53.ListHostedZonesByNameInput{
			DNSName: &domainName,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not retrieve hosted zones")
		return
	}
	if allZones.HostedZones[0].Name != &domainName {
		log.Error().Msg("Could not find " + domainName)
		return
	}
	r.domainID = allZones.HostedZones[0].Id

	// Finding record ID
	allRecords, err := r.client.ListResourceRecordSets(
		r.ctx,
		&route53.ListResourceRecordSetsInput{
			HostedZoneId:    r.domainID,
			StartRecordName: &subDomainName,
			StartRecordType: types.RRTypeA,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not retrieve hosted zone records")
		return
	}
	if allRecords.ResourceRecordSets[0].Name != &subDomainName {
		log.Error().Msg("Could not find " + subDomainName)
		return
	}

	return
}

func (r *R53) Update(ip string) (err error) {
	r.recordSet.ResourceRecords[0].Value = &ip

	_, err = r.client.ChangeResourceRecordSets(
		r.ctx,
		&route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action:            types.ChangeActionUpsert,
						ResourceRecordSet: r.recordSet,
					},
				},
			},
			HostedZoneId: r.domainID,
		},
	)

	return
}
