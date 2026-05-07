/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/api"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for plugins in the EndGit registry",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Searching for \"%s\"...", query)
		s.Start()

		client := api.NewClient()
		response, err := client.GetPlugins(query)

		s.Stop()

		if err != nil {
			color.Red("Search failed: %v", err)
			return
		}

		plugins := response.Data.Plugins

		if len(plugins) == 0 {
			color.Yellow("No plugins found matching \"%s\".", query)
			return
		}

		color.Cyan("\nFound %d plugin(s):\n", len(plugins))
		color.HiBlack(strings.Repeat("-", 80))

		for _, p := range plugins {
			// type badge
			typeStr := ""
			if strings.ToUpper(p.PluginType) == "PYTHON" {
				typeStr = color.HiGreenString("[Py]")
			} else {
				typeStr = color.HiBlueString("[C++]")
			}

			nameStr := color.New(color.FgHiWhite, color.Bold).
				Sprintf("%-20s", p.Name)

			versionStr := color.MagentaString("v%s", fallbackVersion(p.LatestVersion))
			downloadsStr := color.YellowString("%d ⬇", p.Downloads)

			fmt.Printf("%s %s | %s | %s\n",
				typeStr,
				nameStr,
				versionStr,
				downloadsStr,
			)

			fmt.Printf("%s\n", color.HiBlackString(p.Description))
		}

		color.HiBlack(strings.Repeat("-", 80))
		color.White("\nRun %s to install.\n",
			color.HiWhiteString("endgit install <plugin-name>"),
		)
	},
}

func fallbackVersion(v string) string {
	if v == "" {
		return "?.?.?"
	}
	return v
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
