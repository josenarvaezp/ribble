package driver

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

var (
	// ribble role creation variables
	assumeRolePolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": "arn:aws:iam::%s:user/%s",
					"Service": [
						"lambda.amazonaws.com"
					]
				},
				"Action": "sts:AssumeRole",
				"Condition": {}
			}
		]
	}`
	ribbleRoleName        = "ribble"
	ribbleRoleDescription = "Role for ribble operations"
	twelveHrsInSeconds    = int32(43200)
)

var (
	// ribble role policy creation variables
	ribblePolicy = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:*",
				"Resource": "arn:aws:s3::*:*"
			},
			{
				"Effect": "Allow",
				"Action": "sqs:*",
				"Resource": "arn:aws:sqs::*:*"
			},
			{
				"Effect": "Allow",
				"Action": "lambda:*",
				"Resource": "arn:aws:lambda::*:*"
			},
			{
				"Effect": "Allow",
				"Action": "ecr:CreateRepository",
				"Resource": "arn:aws:ecr::*:*"
			}
		]
	}`

	ribblePolicyName        = "ribblePolicy"
	ribblePolicyDescription = "Policy for ribble jobs"
)

var (
	// user policy to assume role variables
	assumeRoleUserPolicy = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "sts:AssumeRole",
				"Resource": "arn:aws:iam::%s:role/ribble"
			}
		]
	}`
	assumeRoleUserPolicyName        = "assumeRibblePolicy"
	assumeRoleUserPolicyDescription = "Policy for users to assume ribble role"
)

// CreateRole creates the ribble role and returns the role ARN
func (d *Driver) CreateRole(ctx context.Context) (*string, error) {
	// check if role exists
	res, err := d.IamAPI.GetRole(ctx, &iam.GetRoleInput{
		RoleName: &ribbleRoleName,
	})
	if err != nil {
		if resourceNotExists(err) {
			// assumeRolePolicy determines who is allowed to use the role
			assumeRolePolicy := fmt.Sprintf(assumeRolePolicyDocument, d.Config.AccountID, d.Config.Username)
			role, err := d.IamAPI.CreateRole(ctx, &iam.CreateRoleInput{
				AssumeRolePolicyDocument: &assumeRolePolicy,
				RoleName:                 &ribbleRoleName,
				Description:              &ribbleRoleDescription,
				MaxSessionDuration:       &twelveHrsInSeconds,
			})
			if err != nil {
				return nil, err
			}

			return role.Role.Arn, nil
		} else {
			return nil, err
		}
	}

	return res.Role.Arn, nil
}

// CreateRolePolicy creates a policy for the ribble role and returns the policy ARN
func (d *Driver) CreateRolePolicy(ctx context.Context) (*string, error) {
	// check if policy exists
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", d.Config.AccountID, ribblePolicyName)
	res, err := d.IamAPI.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: &policyArn,
	})
	if err != nil {
		if resourceNotExists(err) {
			output, err := d.IamAPI.CreatePolicy(ctx, &iam.CreatePolicyInput{
				PolicyDocument: &ribblePolicy,
				PolicyName:     &ribblePolicyName,
				Description:    &ribblePolicyDescription,
			})
			if err != nil {
				return nil, err
			}

			return output.Policy.Arn, nil
		} else {
			return nil, err
		}
	}

	return res.Policy.Arn, nil
}

// AttachPolicy attach policy attaches the ribble policy to the ribble role
func (d *Driver) AttachRolePolicy(ctx context.Context, policyARN *string) error {
	// check if policy is attached to role
	_, err := d.IamAPI.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
		PolicyName: &ribblePolicyName,
		RoleName:   &ribbleRoleName,
	})
	if err != nil {
		if resourceNotExists(err) {
			_, err := d.IamAPI.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
				PolicyArn: policyARN,
				RoleName:  &ribbleRoleName,
			})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// CreateUserPolicy creates a policy for the ribble role and returns the policy ARN
func (d *Driver) CreateUserPolicy(ctx context.Context) (*string, error) {
	// check if user policy exists
	assumePolicyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", d.Config.AccountID, assumeRoleUserPolicyName)
	res, err := d.IamAPI.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: &assumePolicyArn,
	})
	if err != nil {
		if resourceNotExists(err) {
			assumeRoleUserPolicy := fmt.Sprintf(assumeRoleUserPolicy, d.Config.AccountID)
			output, err := d.IamAPI.CreatePolicy(ctx, &iam.CreatePolicyInput{
				PolicyDocument: &assumeRoleUserPolicy,
				PolicyName:     &assumeRoleUserPolicyName,
				Description:    &assumeRoleUserPolicyDescription,
			})
			if err != nil {
				return nil, err
			}
			return output.Policy.Arn, nil
		} else {
			return nil, err
		}
	}

	return res.Policy.Arn, nil
}

// AttachUserPolicy enables the user to assume the ribble role
func (d *Driver) AttachUserPolicy(ctx context.Context, policyARN *string) error {
	// check if user policy is attached to user
	_, err := d.IamAPI.GetUserPolicy(ctx, &iam.GetUserPolicyInput{
		PolicyName: &assumeRoleUserPolicyName,
		UserName:   &d.Config.Username,
	})
	if err != nil {
		if resourceNotExists(err) {
			_, err := d.IamAPI.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
				PolicyArn: policyARN,
				UserName:  &d.Config.Username,
			})
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

// resourceNotExists checks if an iam error is NoSuchEntityException
// to check if resource exists
func resourceNotExists(err error) bool {
	var noSuchEntity *types.NoSuchEntityException
	return errors.As(err, &noSuchEntity)
}
