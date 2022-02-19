package driver

import (
	"context"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/josenarvaezp/displ/internal/aggregators"
)

const (
	scriptToUploadImages = "./build/upload_images.sh"
)

// CreateRepo creates a repository in ECR to upload an image
func (d *Driver) CreateRepo(ctx context.Context, repoName string) error {
	// create repo
	params := &ecr.CreateRepositoryInput{
		RepositoryName: &repoName,
		RegistryId:     &d.Config.AccountID,
	}
	_, err := d.ImageRepoAPI.CreateRepository(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// UploadMapper upploads a mapper image to ECR
func (d *Driver) UploadMapper(ctx context.Context) error {
	err := d.CreateRepo(ctx, d.BuildData.MapperData.ImageName)
	if err != nil {
		return err
	}

	// tag and push image
	_, err = exec.Command(
		scriptToUploadImages,
		d.BuildData.MapperData.ImageName,
		d.BuildData.MapperData.ImageTag,
		d.Config.AccountID,
		d.Config.Region,
	).Output()
	if err != nil {
		return err
	}

	return nil
}

// UploadCoordinator upploads a coordinator image to ECR
func (d *Driver) UploadCoordinator(ctx context.Context) error {
	err := d.CreateRepo(ctx, d.BuildData.CoordinatorData.ImageName)
	if err != nil {
		return err
	}

	// tag and push image
	_, err = exec.Command(
		scriptToUploadImages,
		d.BuildData.CoordinatorData.ImageName,
		d.BuildData.CoordinatorData.ImageTag,
		d.Config.AccountID,
		d.Config.Region,
	).Output()
	if err != nil {
		return err
	}

	return nil
}

// UploadAggregators upploads the aggregator images to ECR
func (d *Driver) UploadAggregators(ctx context.Context) error {
	for _, aggregator := range aggregators.Aggregators {
		err := d.CreateRepo(ctx, aggregator)
		if err != nil {
			return err
		}

		// tag and push image
		_, err = exec.Command(
			scriptToUploadImages,
			aggregator,
			"latest",
			d.Config.AccountID,
			d.Config.Region,
		).Output()
		if err != nil {
			return err
		}
	}

	return nil
}

// UploadJobImages upploads the map, coordinator and
// aggregator images needed for the job
func (d *Driver) UploadJobImages(ctx context.Context) error {
	err := d.UploadMapper(ctx)
	if err != nil {
		return err
	}

	err = d.UploadCoordinator(ctx)
	if err != nil {
		return err
	}

	err = d.UploadAggregators(ctx)
	if err != nil {
		return err
	}

	return nil
}
