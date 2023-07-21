package secrets

import (
	"fmt"

	iaws "github.com/ConradKurth/gokit/aws"
	"github.com/ConradKurth/gokit/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type Secrets struct {
	ssm ssmiface.SSMAPI
}

func New(awsSession *iaws.AWS) *Secrets {
	return &Secrets{
		ssm.New(awsSession.GetSession(), aws.NewConfig().WithRegion(iaws.AWS_REGION)),
	}
}

// LoadSecrets will load secrets from our parameter store
func (s *Secrets) LoadSecrets(c *config.Config) error {
	ignored := map[string]struct{}{}

	mapping := c.GetStringMapString("secrets")
	for name, value := range mapping {
		if _, ok := ignored[name]; ok {
			continue
		}
		if err := s.injectSecret(name, value, c); err != nil {
			return fmt.Errorf("injecting secret for key '%s': %w", name, err)
		}
	}
	return nil
}

// injectSecret will inject secrets into our config for a specified key.
func (s *Secrets) injectSecret(key, path string, c *config.Config) error {
	param, err := s.ssm.GetParameter(&ssm.GetParameterInput{
		Name:           &key,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return err
	}
	c.SetValue(path, *param.Parameter.Value)
	return nil
}
