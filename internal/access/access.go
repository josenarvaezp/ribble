package access

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// IamAPI is an interface used to mock API calls made to the aws IAM service
type IamAPI interface {
	CreateRole(
		ctx context.Context,
		params *iam.CreateRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateRoleOutput, error)
	CreatePolicy(
		ctx context.Context,
		params *iam.CreatePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.CreatePolicyOutput, error)
	AttachRolePolicy(
		ctx context.Context,
		params *iam.AttachRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachRolePolicyOutput, error)
	AttachUserPolicy(
		ctx context.Context,
		params *iam.AttachUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachUserPolicyOutput, error)
	GetRole(
		ctx context.Context,
		params *iam.GetRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.GetRoleOutput, error)
	GetPolicy(
		ctx context.Context,
		params *iam.GetPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetPolicyOutput, error)
	GetRolePolicy(
		ctx context.Context,
		params *iam.GetRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetRolePolicyOutput, error)
	GetUserPolicy(
		ctx context.Context,
		params *iam.GetUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetUserPolicyOutput, error)
}
