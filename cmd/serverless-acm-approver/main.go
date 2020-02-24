package main

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/serverless-acm-approver/pkg/approver"
)

var certApprover approver.Certificate

func createAndApproveACMCertificate(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {

	data := map[string]interface{}{}

	domainName, ok := event.ResourceProperties["DomainName"].(string)
	if !ok {
		return "", nil, errors.New("missing required DomainName from Properties")
	}

	hostedZoneID, ok := event.ResourceProperties["HostedZoneId"].(string)
	if !ok {
		return "", nil, errors.New("missing required HostedZoneId from Properties")
	}

	subjectAlternativeNamesRaw, ok := event.ResourceProperties["SubjectAlternativeNames"].(string)
	if !ok {
		return "", nil, errors.New("missing required SubjectAlternativeNames from Properties")
	}

	subjectAlternativeNames := []string{}

	// avoid splitting empty strings as it results in [""]
	if subjectAlternativeNamesRaw != "" {
		subjectAlternativeNames = strings.Split(subjectAlternativeNamesRaw, ",")
	}

	switch event.RequestType {
	case cfn.RequestDelete:
		err := certApprover.Delete(ctx, event.PhysicalResourceID)
		if err != nil {
			return event.PhysicalResourceID, data, err
		}

		return event.PhysicalResourceID, data, nil
	case cfn.RequestCreate:
		certificateARN, err := certApprover.Request(ctx, event.RequestID, domainName, subjectAlternativeNames, hostedZoneID)
		if err != nil {
			return "", data, err
		}

		return certificateARN, data, nil
	default:
		log.Warn().Str("RequestType", string(event.RequestType)).Str("RequestID", event.RequestID).Msg("no handler for event")
		return event.PhysicalResourceID, data, nil
	}

}

func main() {

	log.Info().Msg("starting lambda")

	certApprover = approver.New()

	lambda.Start(cfn.LambdaWrap(createAndApproveACMCertificate))
}
