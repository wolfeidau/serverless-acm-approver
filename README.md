# serverless-acm-approver

This serverless application provides an acm approver function which uses route53 to aid in the automated creation of an acm certificate.

[![GitHub Actions status](https://github.com/wolfeidau/serverless-acm-approver/workflows/Go/badge.svg?branch=master)](https://github.com/wolfeidau/serverless-acm-approver/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/github.com/wolfeidau/serverless-acm-approver)](https://goreportcard.com/report/github.com/wolfeidau/serverless-acm-approver)
[![Documentation](https://godoc.org/github.com/wolfeidau/serverless-acm-approver?status.svg)](https://godoc.org/github.com/wolfeidau/serverless-acm-approver)

# Why?

The approvers I have used in the past were either limited to creation only, or rather limited in their monitoring / reporting of errors.

This is heavily inspired by the acm [approver lambda](https://github.com/aws/aws-cdk/blob/master/packages/%40aws-cdk/aws-certificatemanager/lambda-packages/dns_validated_certificate_handler/lib/index.js) which is packaged with [AWS CDK](https://github.com/aws/aws-cdk).

Also lots of ideas came from [b-b3rn4rd/acm-approver-lambda](https://github.com/b-b3rn4rd/acm-approver-lambda).

# Usage

The following template illustrates how to use this [serverless application](https://serverlessrepo.aws.amazon.com/applications/arn:aws:serverlessrepo:us-east-1:170889777468:applications~serverless-acm-approver).

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: 'AWS::Serverless-2016-10-31'
Description: >-
  This template demonstrates how to use the serverless-acm-approver application.

Parameters:
  DomainName:
    Type: String
  HostedZoneId:
    Type: String
  SubjectAlternativeNames:
    Type: CommaDelimitedList

Resources:
  ServerlessACMApprover:
    Type: 'AWS::Serverless::Application'
    Properties:
      Location:
        ApplicationId: arn:aws:serverlessrepo:us-east-1:170889777468:applications/serverless-acm-approver
        SemanticVersion: 1.1.0
      Parameters:
        DomainName: !Ref DomainName
        HostedZoneId: !Ref HostedZoneId
        SubjectAlternativeNames:
          !Join
            - ","
            - Ref: SubjectAlternativeNames
       # Optional region to enable creation of ACM certificates in us-east-1 for cloudfront...
       # Region: us-east-1 

Outputs:
  CertificateArn:
    Description: "Certificate ARN"
    Value: !GetAtt ServerlessACMApprover.Outputs.CertificateArn
```

# License

This application is released under Apache 2.0 license and is copyright Mark Wolfe.
