# Dynamic AWS DNS Services

This will keep the supported AWS services up to date with your current public facing IP.
Designed to be run from a docker container or local service to keep an asset up to date.

## Configuration
This is primarily done through a JSON config file (example found in /example/config.json)
It is assumed that AWS credentials will be located in `~/.aws/credentials` with any necessary configurations in `~/.aws/config`

## Currently supported Services, and their required information:
- Route 53
  - Config:
    - Domain Name for hosted zone
    - Subdomain name for target record
  - AWS
    - Profile name: `DynRoute53`
    - IAM Actions:
      - `route53:ListHostedZonesByName`
      - `route53:ListResourceRecordSets`
      - `route53:ChangeResourceRecordSets`
- EC2 - Security Groups
  - Config: 
    - Security Group ID
  - AWS
    - Profile name: `DynSecurityGroup`
    - IAM Actions:
      - `ec2:DescribeSecurityGroupRules`
      - `ec2:ModifySecurityGroupRules`
