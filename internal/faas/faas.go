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
	CreateFunction(
		ctx context.Context,
		params *lambda.CreateFunctionInput,
		optFns ...func(*lambda.Options),
	) (*lambda.CreateFunctionOutput, error)
	PutFunctionConcurrency(
		ctx context.Context,
		params *lambda.PutFunctionConcurrencyInput,
		optFns ...func(*lambda.Options),
	) (*lambda.PutFunctionConcurrencyOutput, error)
	PutFunctionEventInvokeConfig(
		ctx context.Context,
		params *lambda.PutFunctionEventInvokeConfigInput,
		optFns ...func(*lambda.Options),
	) (*lambda.PutFunctionEventInvokeConfigOutput, error)
	PutProvisionedConcurrencyConfig(
		ctx context.Context,
		params *lambda.PutProvisionedConcurrencyConfigInput,
		optFns ...func(*lambda.Options),
	) (*lambda.PutProvisionedConcurrencyConfigOutput, error)
}
