/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/

package update

import (
	"fmt"
	"os"
	"os/exec"
)

func runUpdateScript() error {
	installScriptURL := "https://raw.githubusercontent.com/MagdielCAS/magi-cli/main/scripts/install.sh"
	cmdStr := fmt.Sprintf("curl -sSL %s | bash", installScriptURL)

	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
