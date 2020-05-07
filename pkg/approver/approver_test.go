package approver

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/wolfeidau/serverless-acm-approver/mocks"
)

func TestDelete(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acmapi := mocks.NewMockACMAPI(ctrl)

	acmapi.EXPECT().DescribeCertificateWithContext(gomock.Any(), gomock.Any()).Return(&acm.DescribeCertificateOutput{Certificate: &acm.CertificateDetail{InUseBy: []*string{}}}, nil)
	acmapi.EXPECT().DeleteCertificateWithContext(gomock.Any(), &acm.DeleteCertificateInput{CertificateArn: aws.String("ghi789")}).Return(&acm.DeleteCertificateOutput{}, nil)

	ca := certificateApprover{acm: acmapi}

	err := ca.Delete(context.TODO(), "ghi789")
	assert.NoError(err)
}

func TestApprove(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acmapi := mocks.NewMockACMAPI(ctrl)
	route53api := mocks.NewMockRoute53API(ctrl)

	acmapi.EXPECT().DescribeCertificateWithContext(gomock.Any(), &acm.DescribeCertificateInput{CertificateArn: aws.String("ghi789")}).Return(
		&acm.DescribeCertificateOutput{Certificate: &acm.CertificateDetail{
			CertificateArn: aws.String("ghi789"),
			DomainValidationOptions: []*acm.DomainValidation{
				{
					ResourceRecord: &acm.ResourceRecord{Name: aws.String("_a.1.t.co"), Type: aws.String("CNAME"), Value: aws.String("abc")},
				},
			}}}, nil)

	route53api.EXPECT().ChangeResourceRecordSetsWithContext(gomock.Any(), gomock.Any()).Return(&route53.ChangeResourceRecordSetsOutput{}, nil)
	acmapi.EXPECT().WaitUntilCertificateValidatedWithContext(gomock.Any(), &acm.DescribeCertificateInput{CertificateArn: aws.String("ghi789")}, gomock.Any(), gomock.Any()).Return(nil)

	ca := certificateApprover{acm: acmapi, route53: route53api}

	err := ca.Approve(context.TODO(), "ghi789", "a.1.t.co")
	assert.NoError(err)
}

func TestCreate(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acmapi := mocks.NewMockACMAPI(ctrl)
	route53api := mocks.NewMockRoute53API(ctrl)

	acmapi.EXPECT().RequestCertificateWithContext(gomock.Any(), &acm.RequestCertificateInput{
		DomainName:              aws.String("a.1.t.co"),
		IdempotencyToken:        aws.String("5c69bb695cc29b93d655e1a4bb5656cd"),
		SubjectAlternativeNames: []*string{aws.String("")},
		ValidationMethod:        aws.String("DNS"),
	}).Return(&acm.RequestCertificateOutput{CertificateArn: aws.String("ghi789")}, nil)

	ca := certificateApprover{acm: acmapi, route53: route53api}

	certificateArn, err := ca.Request(context.TODO(), "abc123", "a.1.t.co", []string{""})
	assert.NoError(err)
	assert.Equal("ghi789", certificateArn)
}
