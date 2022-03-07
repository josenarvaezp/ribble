package lambdas

// NewSumReducer initializes a new reducer with its required clients
// func NewSumReducer(
// 	local bool,
// ) (*Reducer, error) {
// 	var cfg *aws.Config
// 	var err error

// 	// get region from env var
// 	region := os.Getenv("AWS_REGION")

// 	// init mapper
// 	reducer := &Reducer{
// 		Region:       region,
// 		Local:        local,
// 		Output:       aggregators.Sum(0),
// 		DedupeSimple: InitDedupeSimple(),
// 	}

// 	// create config
// 	if local {
// 		cfg, err = config.InitLocalCfg()
// 		if err != nil {
// 			return nil, err
// 		}
// 	} else {
// 		cfg, err = config.InitCfg(region)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	// Create an S3 client using the loaded configuration
// 	s3Client := s3.NewFromConfig(*cfg, func(o *s3.Options) {
// 		o.UsePathStyle = true
// 	})
// 	reducer.ObjectStoreAPI = s3Client

// 	// Create a S3 downloader and uploader
// 	reducer.DownloaderAPI = manager.NewDownloader(s3Client)
// 	reducer.UploaderAPI = manager.NewUploader(s3Client)

// 	// create sqs client
// 	reducer.QueuesAPI = sqs.NewFromConfig(*cfg)

// 	return reducer, err
// }