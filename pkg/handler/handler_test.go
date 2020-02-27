package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/golang/mock/gomock"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/serverless-acm-approver/mocks"
)

var paramsJSON = `
{
    "DomainName": "t.1.co",
    "HostedZoneId": "QA8Q",
    "ServiceToken": "arn",
    "SubjectAlternativeNames": [
        ""
    ]
}`

func TestDecodeJSON(t *testing.T) {
	assert := require.New(t)

	mapParams := map[string]interface{}{}
	err := json.Unmarshal([]byte(paramsJSON), &mapParams)
	assert.NoError(err)

	params := new(Params)

	err = mapstructure.Decode(mapParams, params)
	assert.NoError(err)

	assert.Equal(&Params{DomainName: "t.1.co", HostedZoneId: "QA8Q", ServiceToken: "arn", SubjectAlternativeNames: []string{""}}, params)

}

func TestValidate(t *testing.T) {
	assert := require.New(t)

	params := &Params{DomainName: "t.1.co", HostedZoneId: "QA8Q", ServiceToken: "arn", SubjectAlternativeNames: []string{""}}

	assert.NoError(params.Validate())
}

func TestParams_Validate(t *testing.T) {
	type fields struct {
		DomainName              string
		HostedZoneId            string
		ServiceToken            string
		SubjectAlternativeNames []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "validate with good input should return no error",
			fields: fields{
				DomainName:              "t.1.co",
				HostedZoneId:            "QA8Q",
				ServiceToken:            "arn",
				SubjectAlternativeNames: []string{""},
			},
		},
		{
			name: "validate with missing domain name should return error",
			fields: fields{
				HostedZoneId:            "QA8Q",
				ServiceToken:            "arn",
				SubjectAlternativeNames: []string{""},
			},
			wantErr: true,
		},
		{
			name: "validate with missing service token should return error",
			fields: fields{
				DomainName:              "t.1.co",
				HostedZoneId:            "QA8Q",
				SubjectAlternativeNames: []string{""},
			},
			wantErr: true,
		},
		{
			name: "validate with missing HostedZoneId should return error",
			fields: fields{
				DomainName:              "t.1.co",
				ServiceToken:            "arn",
				SubjectAlternativeNames: []string{"", "b.l.co"},
			},
			wantErr: true,
		},
		{
			name: "validate with missing SubjectAlternativeNames should return error",
			fields: fields{
				DomainName:   "t.1.co",
				HostedZoneId: "QA8Q",
				ServiceToken: "arn",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Params{
				DomainName:              tt.fields.DomainName,
				HostedZoneId:            tt.fields.HostedZoneId,
				ServiceToken:            tt.fields.ServiceToken,
				SubjectAlternativeNames: tt.fields.SubjectAlternativeNames,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Params.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCertRequestCreate(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cert := mocks.NewMockCertificate(ctrl)

	cert.EXPECT().Request(gomock.Any(), "abc123", "t.1.co", []string{}, "QA8Q").Return("ghi789", nil)

	dispatcher := &Dispatcher{certApprover: cert}

	event := cfn.Event{
		RequestID:          "abc123",
		PhysicalResourceID: "cde456",
		RequestType:        cfn.RequestCreate,
		ResourceProperties: map[string]interface{}{
			"DomainName":              "t.1.co",
			"HostedZoneId":            "QA8Q",
			"ServiceToken":            "arn",
			"SubjectAlternativeNames": []string{""},
		},
	}

	physicalID, data, err := dispatcher.CreateAndApproveACMCertificate(context.TODO(), event)
	assert.NoError(err)
	assert.Equal("ghi789", physicalID)
	assert.NotNil(data)
}

func TestCertRequestDelete(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cert := mocks.NewMockCertificate(ctrl)

	cert.EXPECT().Delete(gomock.Any(), "ghi789").Return(nil)

	dispatcher := &Dispatcher{certApprover: cert}

	event := cfn.Event{
		RequestID:          "abc123",
		PhysicalResourceID: "ghi789",
		RequestType:        cfn.RequestDelete,
		ResourceProperties: map[string]interface{}{
			"DomainName":              "t.1.co",
			"HostedZoneId":            "QA8Q",
			"ServiceToken":            "arn",
			"SubjectAlternativeNames": []string{""},
		},
	}

	physicalID, data, err := dispatcher.CreateAndApproveACMCertificate(context.TODO(), event)
	assert.NoError(err)
	assert.Equal("ghi789", physicalID)
	assert.NotNil(data)
}
