/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setup(t *testing.T) func() {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("/tmp", "magi-cli-test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("/tmp", 0755); err != nil {
		pterm.Error.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	viper.AddConfigPath("/tmp")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(tmpfile.Name())

	// Return a cleanup function
	return func() {
		os.Remove(tmpfile.Name())
		viper.Reset()
	}
}

func execute(t *testing.T, cmd *cobra.Command, args ...string) string {
	buf := new(bytes.Buffer)

	// Redirect pterm output to the buffer
	pterm.EnableOutput()
	pterm.SetDefaultOutput(buf)
	pterm.Success = *pterm.Success.WithWriter(buf)
	pterm.Error = *pterm.Error.WithWriter(buf)
	pterm.Info = *pterm.Info.WithWriter(buf)
	pterm.DefaultTable = *pterm.DefaultTable.WithWriter(buf)
	// Restore the original output at the end of the test
	defer pterm.SetDefaultOutput(os.Stdout)

	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("error executing command: %v", err)
	}
	return buf.String()
}
