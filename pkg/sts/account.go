package sts

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func GetAccountID(profile, region string) (string, error) {
	// Create a new session
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// Create a new STS client
	stsSvc := sts.New(sess)

	// Call GetCallerIdentity
	input := &sts.GetCallerIdentityInput{}
	result, err := stsSvc.GetCallerIdentity(input)
	if err != nil {
		log.Fatalf("Failed to get caller identity: %v", err)
	}

	return aws.StringValue(result.Account), nil
}
