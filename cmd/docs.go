package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"

	cfg "github.com/cloudposse/atmos/pkg/config"
	"github.com/cloudposse/atmos/pkg/schema"
	u "github.com/cloudposse/atmos/pkg/utils"
)

const atmosDocsURL = "https://atmos.tools"

// docsCmd opens the Atmos docs and can display component documentation
var docsCmd = &cobra.Command{
	Use:                "docs",
	Short:              "Open the Atmos docs or display component documentation",
	Long:               `This command opens the Atmos docs or displays the documentation for a specified Atmos component.`,
	Example:            "atmos docs vpc",
	Args:               cobra.MaximumNArgs(1),
	FParseErrWhitelist: struct{ UnknownFlags bool }{UnknownFlags: false},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			info := schema.ConfigAndStacksInfo{
				Component:             args[0],
				ComponentFolderPrefix: "terraform",
			}

			cliConfig, err := cfg.InitCliConfig(info, true)
			if err != nil {
				u.LogErrorAndExit(schema.CliConfiguration{}, err)
			}

			// Construct the full path to the Terraform component by combining the Atmos base path, Terraform base path, and component name
			componentPath := path.Join(cliConfig.BasePath, cliConfig.Components.Terraform.BasePath, info.Component)

			componentPathExists, err := u.IsDirectory(componentPath)
			if err != nil {
				u.LogErrorAndExit(schema.CliConfiguration{}, err)
			}
			if !componentPathExists {
				u.LogErrorAndExit(schema.CliConfiguration{}, fmt.Errorf("Component '%s' not found in path: '%s'",
					info.Component,
					componentPath,
				))
			}

			readmePath := path.Join(componentPath, "README.md")
			if _, err := os.Stat(readmePath); err != nil {
				if os.IsNotExist(err) {
					u.LogErrorAndExit(schema.CliConfiguration{}, fmt.Errorf("No README found for component: %s", info.Component))
				} else {
					u.LogErrorAndExit(schema.CliConfiguration{}, err)
				}
			}

			readmeContent, err := os.ReadFile(readmePath)
			if err != nil {
				u.LogErrorAndExit(schema.CliConfiguration{}, err)
			}

			componentDocs, err := glamour.Render(string(readmeContent), "dark")
			if err != nil {
				u.LogErrorAndExit(schema.CliConfiguration{}, err)
			}

			fmt.Println(componentDocs)
			return
		}

		// Opens atmos.tools docs if no component argument is provided
		var err error
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", atmosDocsURL).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", atmosDocsURL).Start()
		case "darwin":
			err = exec.Command("open", atmosDocsURL).Start()
		default:
			err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
		}

		if err != nil {
			u.LogErrorAndExit(schema.CliConfiguration{}, err)
		}
	},
}

func init() {
	RootCmd.AddCommand(docsCmd)
}
