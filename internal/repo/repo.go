package repo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// ImageRepoAPI is an interface used to mock API calls made to the aws SQS service
type ImageRepoAPI interface {
	CreateRepository(ctx context.Context,
		params *ecr.CreateRepositoryInput,
		optFns ...func(*ecr.Options),
	) (*ecr.CreateRepositoryOutput, error)
}
