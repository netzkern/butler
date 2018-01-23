package updater

import (
	"fmt"
	"log"
	"os"

	logy "github.com/apex/log"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

func ConfirmAndSelfUpdate(repository string, currentVersion string) {
	latest, found, err := selfupdate.DetectLatest(repository)
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return
	}

	fmt.Printf("Version: %s\n", currentVersion)
	v := semver.MustParse(currentVersion)
	if !found || latest.Version.Equals(v) {
		log.Println("Current version is the latest")
		return
	}

	update := false
	prompt := &survey.Confirm{
		Message: "Do you want to update to " + latest.Version.String() + "?",
	}
	survey.AskOne(prompt, &update, nil)

	if !update {
		return
	}

	cmdPath, err := os.Executable()
	if err != nil {
		logy.WithError(err).Error("os executable")
		return
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, cmdPath); err != nil {
		log.Println("Error occurred while updating binary:", err)
		return
	}
	log.Println("Successfully updated to version", latest.Version)
}
