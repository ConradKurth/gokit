package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// AWS_REGION is the region that this app runs in
const AWS_REGION = "us-west-2"

type AWS struct {
	sess *session.Session
}

func New() *AWS {
	opt := session.Options{
		Config: aws.Config{Region: aws.String(AWS_REGION)},
	}

	profile := os.Getenv("AWS_SSO_PROFILE")
	if profile != "" {
		opt.SharedConfigState = session.SharedConfigEnable
		opt.Profile = profile
	}

	return newSession(opt)
}

func newSession(opt session.Options) *AWS {
	sess, err := session.NewSessionWithOptions(opt)
	if err != nil {
		panic(err)
	}

	return &AWS{
		sess,
	}
}

func (a *AWS) GetSession() *session.Session {
	return a.sess
}
