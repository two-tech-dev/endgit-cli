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
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for plugins in the EndGit registry",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Searching for \"%s\"...", query)
		s.Start()

		client := api.NewClient()
		response, err := client.GetPlugins(query)

		s.Stop()

		if err != nil {
			log.Fatal("Search failed", err)
		}

		plugins := response.Data.Plugins

		if len(plugins) == 0 {
			log.Warnf("No plugins found matching \"%s\"", query)
			return
		}

		fmt.Printf("\nFound %d plugin(s):\n\n", len(plugins))

		// Compute column widths from actual data
		maxName := 4
		maxVer := 6
		for _, p := range plugins {
			if len(p.Name) > maxName {
				maxName = len(p.Name)
			}
			if v := "v" + fallbackVersion(p.LatestVersion); len(v) > maxVer {
				maxVer = len(v)
			}
		}

		const typeWidth = 5 // both "[C++]" and "[PY] " are 5 visible chars
		lineWidth := typeWidth + 1 + maxName + 3 + maxVer + 3 + 10
		if lineWidth < 60 {
			lineWidth = 60
		}
		sep := strings.Repeat("─", lineWidth)

		fmt.Println(sep)

		for _, p := range plugins {
			// Type badge — normalise to 5 visible chars
			tag := "[PY] "
			tagStr := color.HiGreenString(tag)
			if strings.ToUpper(p.PluginType) != "PYTHON" {
				tag = "[C++]"
				tagStr = color.HiBlueString(tag)
			}

			// Pad plain strings BEFORE colorizing (avoids ANSI skewing width)
			namePad := fmt.Sprintf("%-*s", maxName, p.Name)
			nameStr := color.New(color.FgHiWhite, color.Bold).Sprint(namePad)

			verPad := fmt.Sprintf("%-*s", maxVer, "v"+fallbackVersion(p.LatestVersion))
			versionStr := color.MagentaString(verPad)

			dlStr := color.YellowString("%d ⬇", p.Downloads)

			fmt.Printf("%s %s │ %s │ %s\n", tagStr, nameStr, versionStr, dlStr)

			if p.Description != "" {
				indent := strings.Repeat(" ", typeWidth+1)
				maxDesc := lineWidth - typeWidth - 1
				desc := p.Description
				if len(desc) > maxDesc {
					desc = desc[:maxDesc-3] + "..."
				}
				fmt.Printf("%s%s\n", indent, color.HiBlackString(desc))
			}
		}

		fmt.Println(sep)
		fmt.Printf("\nRun %s to view details, or %s to install.\n\n",
			color.HiWhiteString("endgit info <plugin>"),
			color.HiWhiteString("endgit install <plugin>"),
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
