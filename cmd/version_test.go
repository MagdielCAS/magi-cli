/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/pterm/pterm"
)

func TestVersionCmd(t *testing.T) {
	rootCmd := GetRootCmd()
	buf := new(bytes.Buffer)

	// Redirect pterm output to the buffer
	pterm.SetDefaultOutput(buf)
	// Restore the original output at the end of the test
	defer pterm.SetDefaultOutput(os.Stdout)

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("error executing version command: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "magi CLI Version:") {
		t.Errorf("expected output to contain version header, but got '%s'", output)
	}
}
