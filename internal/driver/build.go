package driver

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/utils"
)

const (
	// name of directories and flags for binary
	pathToRoot              = "."
	binaryNameToBuildJob    = "gen_job"
	workspaceFlag           = "--workspace"
	jobIdFlag               = "--job-id"
	scriptToGenerateGoFiles = "./build/generate_lambda_files.sh"
	scriptToBuildImages     = "./build/build_dockerfiles.sh"
	GeneratedFilesDir       = "./build/lambda_gen"
)

// BuildJobGenerationBinary generates a binary file that is used to generate
// the go files and dockerfiles needed to run the job
func (d *Driver) BuildJobGenerationBinary() error {
	_, err := exec.Command(scriptToGenerateGoFiles, d.JobID.String(), d.JobPath).Output()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// GenerateResourcesFromBinary runs the binary generated from BuildJobGenerationBinary
// and creates go files and dockerfiles required for the job
func (d *Driver) GenerateResourcesFromBinary() error {
	jobBinaryName := fmt.Sprintf( // ./build/lambda_gen/JOB_ID/gen_job
		"%s/%s/%s",
		GeneratedFilesDir,
		d.JobID.String(),
		binaryNameToBuildJob,
	)
	workspaceSplit := strings.SplitAfterN(d.JobPath, "/", 3)
	workspace := workspaceSplit[1][0 : len(workspaceSplit[1])-1]
	_, err := exec.Command(
		jobBinaryName,
		workspaceFlag,
		workspace,
		jobIdFlag,
		d.JobID.String(),
	).Output()
	if err != nil {
		return err
	}

	return nil
}

// BuildDockerImages is used to generate the images from the generated
// dockerfiles
func (d *Driver) BuildDockerImages() error {
	// build docker files
	dockefilesDir := fmt.Sprintf( // ./build/lambda_gen/JOB_ID/dockerfiles
		"%s/%s/dockerfiles",
		GeneratedFilesDir,
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
			scriptToBuildImages,
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
