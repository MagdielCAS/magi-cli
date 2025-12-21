package utils

import (
	"io"
	"net/http"
	"runtime"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/MagdielCAS/pcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

func GetRepoPath() string {
	return "MagdielCAS/magi-cli"
}

// IsUpdateAvailable checks if there is a newer version available on GitHub.
// It returns true if an update is available, the latest tag name, and any error encountered.
func IsUpdateAvailable(currentVersion string) (bool, string, error) {
	resp, err := http.Get(pterm.Sprintf("https://api.github.com/repos/%s/releases/latest", GetRepoPath()))
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	tagName := gjson.Get(string(body), "tag_name").String()

	if !strings.HasPrefix(tagName, "v") {
		tagName = "v" + tagName
	}

	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}

	if semver.Compare(tagName, currentVersion) > 0 {
		return true, tagName, nil
	}

	return false, tagName, nil
}

// CheckForUpdates checks if a new version of your application is pushed, and notifies the user, if DisableUpdateChecking is true.
func CheckForUpdates(rootCmd *cobra.Command) error {
	if !pcli.DisableUpdateChecking {
		available, tagName, err := IsUpdateAvailable(rootCmd.Version)
		if err != nil {
			return err
		}

		if available {
			format := "A new version of %s is available (%s)!\n"
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
