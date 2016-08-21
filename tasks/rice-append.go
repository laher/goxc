package tasks

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/executils"
)

const riceNotFound = `Could not find 'rice' executable on $PATH.
Please ensure you have installed it:

go get -u github.com/GeertJohan/go.rice/rice

Error: %v`

var riceAppendTask = Task{
	TASK_RICE_APPEND,
	"Append embedded resources to the binary using go.rice.",
	runTaskRiceAppend,
	map[string]interface{}{
		"import-paths": []string{},
	}}

//runs automatically
func init() {
	Register(riceAppendTask)
}

func runTaskRiceAppend(tp TaskParams) (err error) {
	ricePath, err := exec.LookPath("rice")
	if err != nil {
		return fmt.Errorf(riceNotFound, err)
	}
	for _, dest := range tp.DestPlatforms {
		for _, mainDir := range tp.MainDirs {
			var exeName string
			if len(tp.MainDirs) == 1 {
				exeName = tp.Settings.AppName
			} else {
				exeName = filepath.Base(mainDir)

			}
			binPath, err := core.GetAbsoluteBin(dest.Os, dest.Arch, tp.Settings.AppName, exeName, tp.WorkingDirectory, tp.Settings.GetFullVersionName(), tp.Settings.OutPath, tp.Settings.ArtifactsDest)

			if err != nil {
				return err
			}
			if err = riceAppendPlat(dest.Os, dest.Arch, binPath, ricePath, tp.Settings); err != nil {
				return err
			}
		}
	}
	return nil
}

func riceAppendPlat(goos, arch string, binPath string, ricePath string, settings *config.Settings) error {
	importPaths := settings.GetTaskSettingStringSlice("rice-append", "import-paths")
	if err := riceAppend(binPath, ricePath, importPaths); err != nil {
		log.Printf("rice-append failed for %s: %s", binPath, err)
		return err
	}
	if !settings.IsQuiet() {
		log.Printf("rice-append successful for: %s", binPath)
	}
	return nil
}

func riceAppend(binPath string, ricePath string, importPaths []string) error {
	cmd := exec.Command(ricePath)
	cmd.Args = append(cmd.Args, "append", fmt.Sprintf("--exec=%s", binPath))
	for _, importPath := range importPaths {
		cmd.Args = append(cmd.Args, fmt.Sprintf("--import-path=%s", importPath))
	}

	return executils.StartAndWait(cmd)
}
