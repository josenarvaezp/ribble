package lambdas

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/pkg/aggregators"
)

// NewMapReducer initializes a new reducer with its required clients
func NewMapReducer(
	local bool,
) (*Reducer, error) {
	var cfg *aws.Config
	var err error

	// get region from env var
	region := os.Getenv("AWS_REGION")

	// init mapper
	reducer := &Reducer{
		Region: region,
		Local:  local,
		Output: make(aggregators.MapAggregator),
		Dedupe: InitDedupe(),
	}

	// create config
	if local {
		cfg, err = config.InitLocalCfg(region)
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.InitCfg(region)
		if err != nil {
			return nil, err
		}
	}

	// Create an S3 client using the loaded configuration
	s3Client := s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	reducer.ObjectStoreAPI = s3Client

	// Create a S3 downloader and uploader
	reducer.DownloaderAPI = manager.NewDownloader(s3Client)
	reducer.UploaderAPI = manager.NewUploader(s3Client)

	// create sqs client
	reducer.QueuesAPI = sqs.NewFromConfig(*cfg)

	return reducer, err
}
