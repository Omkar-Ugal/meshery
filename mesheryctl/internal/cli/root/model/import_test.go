package model

import (
	"errors"
	"fmt"
	"github.com/jarcoal/httpmock"
	mesheryctlflags "github.com/meshery/meshery/mesheryctl/internal/cli/pkg/flags"
	"github.com/meshery/meshery/mesheryctl/pkg/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestHasCSVs(t *testing.T) {
	tests := []struct {
		name           string
		dirPath        string
		expectedResult bool
	}{
		{
			name:           "directory with CSVs",
			dirPath:        "./fixtures/with_csvs",
			expectedResult: true,
		},
		{
			name:           "directory without CSVs",
			dirPath:        "./fixtures/without_csvs",
			expectedResult: false,
		},
		{
			name:           "non-existent directory",
			dirPath:        "./fixtures/invalid_path",
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := hasCSVs(tc.dirPath)

			assert.Equal(t, tc.expectedResult, res)
		})
	}
}

func TestImportModelReturnsErrorForURLImportFailure(t *testing.T) {
	defer utils.ResetCommandFlags(ModelCmd, t)
	testContext := utils.InitTestEnvironment(t)
	defer utils.StopMockery(t)

	utils.TokenFlag = utils.GetToken(t)
	mesheryctlflags.InitValidators(ModelCmd)

	httpmock.RegisterResponder("POST", testContext.BaseURL+"/api/meshmodels/register",
		httpmock.NewStringResponder(500, "internal server error"))

	buf := utils.SetupMeshkitLoggerTesting(t, false)
	ModelCmd.SetArgs([]string{"import", "https://example.com/model"})
	ModelCmd.SetOut(buf)

	err := ModelCmd.Execute()
	if err == nil {
		t.Fatal("expected an error but command succeeded")
	}

	utils.AssertMeshkitErrorsEqual(t, err, utils.ErrMesheryServerInternalError(errors.New("internal server error")))
}

func TestImportModelDisplaysUsageInstructions(t *testing.T) {
	defer utils.ResetCommandFlags(ModelCmd, t)

	buf := utils.SetupMeshkitLoggerTesting(t, false)
	ModelCmd.SetArgs([]string{"import"})
	ModelCmd.SetOut(buf)

	err := ModelCmd.Execute()
	if err == nil {
		t.Fatal("expected an error but command succeeded")
	}

	expectedErr := utils.ErrInvalidArgument(
		fmt.Errorf("either --file [ file | filepath | URL ] or an argument must be specified\n\n%s", errImportUsageMsg),
	)
	utils.AssertMeshkitErrorsEqual(t, err, expectedErr)
}

func TestImportModelFailsForZipFormat(t *testing.T) {
	defer utils.ResetCommandFlags(ModelCmd, t)

	tmpDir := t.TempDir()
	zipFilePath := filepath.Join(tmpDir, "model.zip")
	if err := os.WriteFile(zipFilePath, []byte("fake zip content"), 0644); err != nil {
		t.Fatalf("failed to create temp zip file: %v", err)
	}

	buf := utils.SetupMeshkitLoggerTesting(t, false)
	ModelCmd.SetArgs([]string{"import", zipFilePath})
	ModelCmd.SetOut(buf)

	err := ModelCmd.Execute()
	if err == nil {
		t.Fatal("expected an error but command succeeded")
	}

	expectedErr := utils.ErrInvalidArgument(
		fmt.Errorf(".zip format is not supported. Please use .tar.gz\n\n%s", errImportUsageMsg),
	)
	utils.AssertMeshkitErrorsEqual(t, err, expectedErr)
}
