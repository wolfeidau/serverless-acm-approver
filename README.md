# serverless-acm-approver

This serverless application provides an acm approver function which uses route53 to aid in the automated creation of an acm certificate.

# Why?

The approvers I have used in the past were either limited to creation only, or rather limited in their monitoring / reporting of errors.

This is heavily inspired by the acm [approver lambda](https://github.com/aws/aws-cdk/blob/master/packages/%40aws-cdk/aws-certificatemanager/lambda-packages/dns_validated_certificate_handler/lib/index.js) which is packaged with [AWS CDK](https://github.com/aws/aws-cdk).

Also lots of ideas came from https://github.com/b-b3rn4rd/acm-approver-lambda.

# License

This application is released under Apache 2.0 license and is copyright Mark Wolfe.