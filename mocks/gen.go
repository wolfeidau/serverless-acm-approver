package mocks

//go:generate env GOBIN=$PWD/bin GO111MODULE=on go install github.com/golang/mock/mockgen
//go:generate $PWD/bin/mockgen -destination=approver.go -package=mocks github.com/wolfeidau/serverless-acm-approver/pkg/approver Certificate
//go:generate $PWD/bin/mockgen -destination acm.go -package=mocks github.com/aws/aws-sdk-go/service/acm/acmiface ACMAPI
//go:generate $PWD/bin/mockgen -destination route53.go -package=mocks github.com/aws/aws-sdk-go/service/route53/route53iface Route53API
