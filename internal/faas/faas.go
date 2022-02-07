package faas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// FaasAPI is an interface used to mock API calls made to the aws Lambda service
type FaasAPI interface {
	Invoke(
		ctx context.Context,
		params *lambda.InvokeInput,
		optFns ...func(*lambda.Options),
	) (*lambda.InvokeOutput, error)
	AddPermission(
		ctx context.Context,
		params *lambda.AddPermissionInput,
		optFns ...func(*lambda.Options),
	) (*lambda.AddPermissionOutput, error)
}
