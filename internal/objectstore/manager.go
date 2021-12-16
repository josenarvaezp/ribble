package objectstore

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ManagerDownloaderAPI is an interface used to mock API calls made to the aws S3 manager downloader
type ManagerDownloaderAPI interface {
	Download(
		ctx context.Context,
		w io.WriterAt,
		input *s3.GetObjectInput,
		options ...func(*manager.Downloader),
	) (n int64, err error)
}

// ManagerUploaderAPI is an interface used to mock API calls made to the aws S3 manager uploader
type ManagerUploaderAPI interface {
	Upload(
		ctx context.Context,
		input *s3.PutObjectInput,
		opts ...func(*manager.Uploader),
	) (*manager.UploadOutput, error)
}
