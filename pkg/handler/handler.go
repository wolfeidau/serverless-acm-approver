package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/serverless-acm-approver/pkg/approver"
)

// Dispatcher dispatches handler requests and holds approver helper
type Dispatcher struct {
	certApprover approver.Certificate
}

// New create a new dispatcher of handlers
func New(config ...*aws.Config) *Dispatcher {
	return &Dispatcher{
		certApprover: approver.New(config...),
	}
}

// Params used to parse inputs to create handler from CFN
type Params struct {
	DomainName              string
	HostedZoneId            string
	ServiceToken            string
	SubjectAlternativeNames []string
	Region                  string
}

// Validate checks the params are valid
func (p *Params) Validate() error {
	if p.DomainName == "" {
		return errors.New("missing required DomainName from Properties")
	}

	if p.ServiceToken == "" {
		return errors.New("missing required ServiceToken from Properties")
	}

	if p.HostedZoneId == "" {
		return errors.New("missing required HostedZoneId from Properties")
	}

	if p.SubjectAlternativeNames == nil {
		return errors.New("missing required SubjectAlternativeNames from Properties")
	}

	filtered := []string{}

	for _, v := range p.SubjectAlternativeNames {
		if v != "" {
			filtered = append(filtered, v)
		}
	}

	p.SubjectAlternativeNames = filtered

	return nil
}

// CreateAndApproveACMCertificate custom cfn certificate creation function
func (ds *Dispatcher) CreateAndApproveACMCertificate(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {

	jsonData, _ := json.Marshal(&event.ResourceProperties)

	fmt.Println(string(jsonData))

	data := map[string]interface{}{}

	params := new(Params)

	err := mapstructure.Decode(event.ResourceProperties, params)
	if err != nil {
		return event.PhysicalResourceID, data, err
	}

	err = params.Validate()
	if err != nil {
		return event.PhysicalResourceID, data, err
	}

	// using the default cert approver to ensure we can test this method
	certApprover := ds.certApprover

	// if a region is passed in then override the client to use it, this is primarily to support
	// targeting us-east-1 for ACM certificates used by cloudfront
	if params.Region != "" {
		certApprover = approver.New(aws.NewConfig().WithRegion(params.Region))
	}

	switch event.RequestType {
	case cfn.RequestDelete:
		err := ds.certApprover.Delete(ctx, event.PhysicalResourceID)
		if err != nil {
			return event.PhysicalResourceID, data, err
		}

		return event.PhysicalResourceID, data, nil
	case cfn.RequestCreate, cfn.RequestUpdate:
		certificateARN, err := certApprover.Request(ctx, event.RequestID, params.DomainName, params.SubjectAlternativeNames, params.HostedZoneId)
		if err != nil {
			return "", data, err
		}

		err = certApprover.Approve(ctx, certificateARN, 300, params.HostedZoneId)
		if err != nil {
			return certificateARN, data, err
		}

		return certificateARN, data, nil
	default:
		log.Warn().Str("RequestType", string(event.RequestType)).Str("RequestID", event.RequestID).Msg("no handler for event")
		return event.PhysicalResourceID, data, nil
	}

}
