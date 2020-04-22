package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/aws/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

//GetCredentials get aws creds thru queryoptions
func GetCredentials(opt *AthenaDatasourceQueryOption) (aws.CredentialsProvider, error) {
	switch opt.AuthType {
	case Static:
		return getStaticCreds(opt)
	case RoleArn:
		return getRoleCreds(opt)
	default:
		return nil, nil
	}
}

func getStaticCreds(opt *AthenaDatasourceQueryOption) (aws.CredentialsProvider, error) {
	sessionToken := ""
	if opt.AccessKey != "" && opt.SecretKey != "" {
		return aws.NewStaticCredentialsProvider(opt.AccessKey, opt.SecretKey, sessionToken), nil
	}
	return nil, nil
}

func getRoleCreds(opt *AthenaDatasourceQueryOption) (aws.CredentialsProvider, error) {
	if opt.RoleARN == "" {
		return nil, nil
	}
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	stsSvc := sts.New(cfg)
	stsCredProvider := stscreds.NewAssumeRoleProvider(stsSvc, string(opt.RoleARN))
	return stsCredProvider, nil
}
