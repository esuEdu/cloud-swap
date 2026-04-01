package credential

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/esuEdu/cloud-swap/internal/domain"
)

func writeAwsFiles(c domain.Credential) error {
	awsDir := filepath.Join(os.Getenv("HOME"), ".aws")
	if err := os.MkdirAll(awsDir, 0700); err != nil {
		return err
	}

	credContent := fmt.Sprintf(
		"[default]\naws_access_key_id=%s\naws_secret_access_key=%s\n",
		c.AccessKey,
		c.SecretKey,
	)

	if c.SessionTok != "" {
		credContent += fmt.Sprintf("aws_session_token=%s\n", c.SessionTok)
	}

	if err := os.WriteFile(filepath.Join(awsDir, "credentials"), []byte(credContent), 0600); err != nil {
		return err
	}

	configContent := fmt.Sprintf(
		"[default]\nregion=%s\noutput=%s\n",
		c.Region,
		c.Output,
	)

	if err := os.WriteFile(filepath.Join(awsDir, "config"), []byte(configContent), 0600); err != nil {
		return err
	}

	return nil
}

func assumeRole(roleArn string, baseCred *domain.Credential) (*domain.Credential, error) {
	creds := credentials.NewStaticCredentials(baseCred.AccessKey, baseCred.SecretKey, baseCred.SessionTok)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(baseCred.Region),
		Credentials: creds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	svc := sts.New(sess)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("cloud-swap-" + time.Now().Format("20060102150405")),
		DurationSeconds: aws.Int64(3600),
	}

	result, err := svc.AssumeRole(input)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	expiresAt := ""
	if result.Credentials != nil && result.Credentials.Expiration != nil {
		expiresAt = result.Credentials.Expiration.Format(time.RFC3339)
	}

	return &domain.Credential{
		Name:       "assumed-" + roleArn,
		Provider:   "aws",
		AccessKey:  *result.Credentials.AccessKeyId,
		SecretKey:  *result.Credentials.SecretAccessKey,
		SessionTok: *result.Credentials.SessionToken,
		ExpiresAt:  expiresAt,
		Region:     baseCred.Region,
		Output:     baseCred.Output,
	}, nil
}

func validateCredential(c *domain.Credential) error {
	if c.AccessKey == "" || c.SecretKey == "" {
		return fmt.Errorf("access_key and secret_key are required")
	}

	if len(c.AccessKey) < 16 {
		return fmt.Errorf("access_key appears invalid (too short)")
	}

	if c.Region == "" {
		return fmt.Errorf("region is required")
	}

	creds := credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, c.SessionTok)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(c.Region),
		Credentials: creds,
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	svc := sts.New(sess)
	_, err = svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("credentials validation failed: %w", err)
	}

	return nil
}
