package utils

import (
	"io"
	"net/http"
	"runtime"

	"github.com/MagdielCAS/pcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

func getRepoPath() string {
	return "MagdielCAS/magi-cli"
}

// CheckForUpdates checks if a new version of your application is pushed, and notifies the user, if DisableUpdateChecking is true.
func CheckForUpdates(rootCmd *cobra.Command) error {
	if !pcli.DisableUpdateChecking {
		resp, err := http.Get(pterm.Sprintf("https://api.github.com/repos/%s/releases/latest", getRepoPath()))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		tagName := gjson.Get(string(body), "tag_name").String()

		if rootCmd.Version != tagName && tagName != "" {
			format := "A new version of %s is availble (%s)!\n"
			format += "You can install the new version with: "

			switch runtime.GOOS {
			case "windows":
				format += pterm.Magenta(pterm.Sprintf(`iwr https://raw.githubusercontent.com/MagdielCAS/magi-cli/main/scripts/install.sh | iex`))
			case "darwin":
				format += pterm.Magenta(pterm.Sprintf(`curl -sSL https://raw.githubusercontent.com/MagdielCAS/magi-cli/main/scripts/install.sh | bash`))
			default:
				format += pterm.Magenta(pterm.Sprintf(`curl -sSL https://raw.githubusercontent.com/MagdielCAS/magi-cli/main/scripts/install.sh | bash`))
			}
			pterm.Info.Printfln(format, rootCmd.Name(), pterm.Magenta(tagName))
		}
	}

	return nil
}
