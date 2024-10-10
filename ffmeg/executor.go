package ffmeg

import (
	"fmt"
	"os"
	"os/exec"
)

const scriptPath = "./ffmeg/create_video.sh"

// ExecuteScript Function to execute the shell script
func ExecuteScript(workDir string) error {
	// Check if the script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script does not exist at: %s", scriptPath)
	}

	// Construct the command to execute the script
	cmd := exec.Command("/bin/bash", scriptPath, workDir)

	// Set the command's output to the standard output and standard error of the current process
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command and capture the error (if any)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute script: %v", err)
	}

	return nil
}
