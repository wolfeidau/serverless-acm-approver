package main

import (
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/serverless-acm-approver/pkg/handler"
)

func main() {

	log.Info().Msg("starting lambda")

	dispatcher := handler.New()

	lambda.Start(cfn.LambdaWrap(dispatcher.CreateAndApproveACMCertificate))
}
