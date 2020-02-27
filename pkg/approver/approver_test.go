package approver

import (
	"context"
	"testing"

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
