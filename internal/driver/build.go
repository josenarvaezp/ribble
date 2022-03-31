package driver

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/generators"
	"github.com/josenarvaezp/displ/internal/utils"
)

// BuildJobGenerationBinary generates a binary file that is used to generate
// the go files and dockerfiles needed to run the job
func (d *Driver) BuildJobGenerationBinary() error {
	_, err := exec.Command(
		generators.ScriptToGenerateGoFiles,
		d.BuildData.BuildDir,
		d.BuildData.JobPath,
	).Output()
	if err != nil {
		return err
	}

	return nil
}

// GenerateResourcesFromBinary runs the binary generated from BuildJobGenerationBinary
// and creates go files and dockerfiles required for the job
func (d *Driver) GenerateResourcesFromBinary() error {
	jobBinaryName := fmt.Sprintf( // ./build/lambda_gen/JOB_ID/gen_job
		"%s/%s",
		d.BuildData.BuildDir,
		generators.BinaryNameToBuildJob,
	)
	workspaceSplit := strings.SplitAfterN(d.BuildData.JobPath, "/", 3)
	workspaceRoot := workspaceSplit[1][0 : len(workspaceSplit[1])-1]
	_, err := exec.Command(
		jobBinaryName,
		generators.WorkspaceFlag,
		workspaceRoot,
		generators.JobIdFlag,
		d.JobID.String(),
	).Output()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) BuildAggregatorImages() error {
	_, err := exec.Command(generators.ScriptToBuildAggregatorImages).Output()
	if err != nil {
		return err
	}

	return nil
}

// BuildDockerImages is used to generate the images from the generated
// dockerfiles
func (d *Driver) BuildDockerCustomImages() error {
	// build docker files
	dockefilesDir := fmt.Sprintf( // ./build/lambda_gen/JOB_ID/dockerfiles
		"%s/%s/dockerfiles",
		generators.GeneratedFilesDir,
		d.JobID.String(),
	)
	err := buildDockerfile(dockefilesDir, d.JobID)
	if err != nil {
		return err
	}

	return nil
}

// buildDockerfile builds the images required to run the job
func buildDockerfile(directory string, JobID uuid.UUID) error {
	files, err := utils.ListFiles(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		// get file name without extension
		imageName := strings.SplitAfter(file.Name(), "Dockerfile.")[1]
		tagName := fmt.Sprintf("%s_%s", strings.ToLower(imageName), JobID.String())
		_, err := exec.Command(
			generators.ScriptToBuildImages,
			tagName,
			JobID.String(),
			imageName,
		).Output()
		if err != nil {
			return err
		}
	}

	return nil
}
