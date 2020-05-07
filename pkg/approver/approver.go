package approver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	maxAttempts          = 20
	recordTTLSeconds     = 60
	describePollWaitTime = 5 * time.Second
	validationPollTime   = 30 * time.Second
	deletionPollTime     = 30 * time.Second
)

// Certificate AWS ACM approver
type Certificate interface {
	Approve(ctx context.Context, certificateArn string, hostedZoneID string) error
	Request(ctx context.Context, requestID string, domainName string, subjectAlternativeNames []string) (string, error)
	Delete(ctx context.Context, certificateArn string) error
}

// Approver the ACM approver
type certificateApprover struct {
	acm     acmiface.ACMAPI
	route53 route53iface.Route53API
}

// New creates a new approver
func New(config ...*aws.Config) Certificate {
	sess := session.Must(session.NewSession(config...))

	return &certificateApprover{
		acm:     acm.New(sess),
		route53: route53.New(sess),
	}
}

func (ac *certificateApprover) Approve(ctx context.Context, certificateArn string, hostedZoneID string) error {
	var (
		err error
		res *acm.DescribeCertificateOutput
	)

	for i := 1; i < maxAttempts; i++ {
		log.Info().Str("certificateArn", certificateArn).Msg("describe certificate")

		res, err = ac.acm.DescribeCertificateWithContext(ctx, &acm.DescribeCertificateInput{
			CertificateArn: aws.String(certificateArn),
		})
		if err != nil {
			return err
		}

		if len(res.Certificate.DomainValidationOptions) > 0 {
			if res.Certificate.DomainValidationOptions[0].ResourceRecord != nil {
				log.Info().Str("certificateArn", certificateArn).Msg("certificate contains confirmation record")
				break
			}
		}

		time.Sleep(describePollWaitTime)
	}

	for _, validation := range res.Certificate.DomainValidationOptions {
		record := validation.ResourceRecord

		log.Info().Msgf("Upserting DNS record into zone %s: %s %s %s",
			hostedZoneID, aws.StringValue(record.Name), aws.StringValue(record.Type), aws.StringValue(record.Value))

		_, err = ac.route53.ChangeResourceRecordSetsWithContext(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(hostedZoneID),
			ChangeBatch: &route53.ChangeBatch{Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: record.Name,
						Type: record.Type,
						TTL:  aws.Int64(recordTTLSeconds),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: record.Value,
							},
						},
					},
				},
			}},
		})
		if err != nil {
			return err
		}
	}

	log.Info().Str("certificateArn", certificateArn).Msg("waiting for certificate validation")

	err = ac.acm.WaitUntilCertificateValidatedWithContext(ctx, &acm.DescribeCertificateInput{
		CertificateArn: res.Certificate.CertificateArn,
	}, request.WithWaiterMaxAttempts(maxAttempts), request.WithWaiterDelay(request.ConstantWaiterDelay(validationPollTime)))
	if err != nil {
		return err
	}

	return nil
}

func (ac *certificateApprover) Request(ctx context.Context, requestID, domainName string, subjectAlternativeNames []string) (string, error) {
	// unique hash of cloudformation request id to ensure only one
	// certificate is created for this CFN request
	token := sum(requestID)

	input := &acm.RequestCertificateInput{
		DomainName:       aws.String(domainName),
		ValidationMethod: aws.String(acm.ValidationMethodDns),
		IdempotencyToken: aws.String(token),
	}

	log.Info().Strs("subjectAlternativeNames", subjectAlternativeNames).Str("token", token).Msg("Request Certificate")

	if len(subjectAlternativeNames) > 0 {
		input.SubjectAlternativeNames = aws.StringSlice(subjectAlternativeNames)
	}

	res, err := ac.acm.RequestCertificateWithContext(ctx, input)
	if err != nil {
		return "", errors.Wrap(err, "failed to Request Certificate")
	}

	certificateArn := aws.StringValue(res.CertificateArn)

	log.Info().Str("arn", certificateArn).Msg("requested certificate")

	return certificateArn, nil
}

func (ac *certificateApprover) Delete(ctx context.Context, certificateArn string) error {
	log.Info().Str("certificateArn", certificateArn).Msg("Delete waiting for InUseBy of 0")

	for i := 1; i < maxAttempts; i++ {
		res, err := ac.acm.DescribeCertificateWithContext(ctx, &acm.DescribeCertificateInput{
			CertificateArn: aws.String(certificateArn),
		})
		if err != nil {
			return err
		}

		if len(res.Certificate.InUseBy) == 0 {
			log.Info().Int("InUseBy", len(res.Certificate.InUseBy)).Msg("certificate InUseBy check done")
			break
		}

		time.Sleep(deletionPollTime)
	}

	log.Info().Str("certificateArn", certificateArn).Msg("deleting certificate")

	_, err := ac.acm.DeleteCertificateWithContext(ctx, &acm.DeleteCertificateInput{
		CertificateArn: aws.String(certificateArn)})
	if err != nil {
		return err
	}

	return nil
}

func sum(requestID string) string {
	data := sha256.Sum224([]byte(requestID))
	return fmt.Sprintf("%x", data[:16])
}
