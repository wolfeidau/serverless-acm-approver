package approver

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/serverless-acm-approver/mocks"
)

func TestDelete(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acmapi := mocks.NewMockACMAPI(ctrl)

	acmapi.EXPECT().DescribeCertificate(gomock.Any()).Return(&acm.DescribeCertificateOutput{Certificate: &acm.CertificateDetail{InUseBy: []*string{}}}, nil)
	acmapi.EXPECT().DeleteCertificate(&acm.DeleteCertificateInput{CertificateArn: aws.String("ghi789")}).Return(&acm.DeleteCertificateOutput{}, nil)

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

	acmapi.EXPECT().DescribeCertificate(&acm.DescribeCertificateInput{CertificateArn: aws.String("ghi789")}).Return(
		&acm.DescribeCertificateOutput{Certificate: &acm.CertificateDetail{
			CertificateArn: aws.String("ghi789"),
			DomainValidationOptions: []*acm.DomainValidation{
				&acm.DomainValidation{
					ResourceRecord: &acm.ResourceRecord{Name: aws.String("_a.1.t.co"), Type: aws.String("CNAME"), Value: aws.String("abc")},
				},
			}}}, nil)

	route53api.EXPECT().ChangeResourceRecordSets(gomock.Any()).Return(&route53.ChangeResourceRecordSetsOutput{}, nil)
	acmapi.EXPECT().WaitUntilCertificateValidatedWithContext(gomock.Any(), &acm.DescribeCertificateInput{CertificateArn: aws.String("ghi789")}, gomock.Any(), gomock.Any()).Return(nil)

	ca := certificateApprover{acm: acmapi, route53: route53api}

	err := ca.Approve(context.TODO(), "ghi789", 300, "a.1.t.co")
	assert.NoError(err)
}

func TestCreate(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acmapi := mocks.NewMockACMAPI(ctrl)
	route53api := mocks.NewMockRoute53API(ctrl)

	acmapi.EXPECT().RequestCertificate(&acm.RequestCertificateInput{
		DomainName:              aws.String("a.1.t.co"),
		IdempotencyToken:        aws.String("e99a18c428cb38d5f260853678922e03"),
		SubjectAlternativeNames: []*string{aws.String("")},
		ValidationMethod:        aws.String("DNS"),
	}).Return(&acm.RequestCertificateOutput{CertificateArn: aws.String("ghi789")}, nil)

	acmapi.EXPECT().DescribeCertificate(&acm.DescribeCertificateInput{CertificateArn: aws.String("ghi789")}).Return(
		&acm.DescribeCertificateOutput{Certificate: &acm.CertificateDetail{
			CertificateArn: aws.String("ghi789"),
			DomainValidationOptions: []*acm.DomainValidation{
				&acm.DomainValidation{
					ResourceRecord: &acm.ResourceRecord{Name: aws.String("_a.1.t.co"), Type: aws.String("CNAME"), Value: aws.String("abc")},
				},
			}}}, nil)

	route53api.EXPECT().ChangeResourceRecordSets(gomock.Any()).Return(&route53.ChangeResourceRecordSetsOutput{}, nil)
	acmapi.EXPECT().WaitUntilCertificateValidatedWithContext(gomock.Any(), &acm.DescribeCertificateInput{CertificateArn: aws.String("ghi789")}, gomock.Any(), gomock.Any()).Return(nil)

	ca := certificateApprover{acm: acmapi, route53: route53api}

	certificateArn, err := ca.Request(context.TODO(), "abc123", "a.1.t.co", []string{""}, "AZ123")
	assert.NoError(err)
	assert.Equal("ghi789", certificateArn)
}
