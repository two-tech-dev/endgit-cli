/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	box "github.com/box-cli-maker/box-cli-maker/v3"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/api"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var searchResultBox = box.NewBox().
			Style(box.Single).
			Padding(1, 0).
			TitlePosition(box.Top).
			ContentAlign(box.Left).
			Color(box.Cyan).
			TitleColor(box.BrightCyan)

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

		var lines []string

		for i, p := range plugins {
			tag := color.HiGreenString("[PY]")
			if strings.ToUpper(p.PluginType) != "PYTHON" {
				tag = color.HiBlueString("[C++]")
			}

			name := color.New(color.FgHiWhite, color.Bold).Sprint(p.Name)
			ver := color.MagentaString("v" + fallbackVersion(p.LatestVersion))
			dl := color.YellowString("%d downloads", p.Downloads)

			lines = append(lines, fmt.Sprintf("%s  %s  %s  %s", tag, name, ver, dl))

			if p.Description != "" {
				desc := p.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				lines = append(lines, "  "+color.HiBlackString(desc))
			}
			if i < len(plugins)-1 {
				lines = append(lines, "")
			}
		}

		content := "\n" + strings.Join(lines, "\n") + "\n"
		title := fmt.Sprintf("Found %d plugin(s)", len(plugins))
		out := searchResultBox.MustRender(title, content)
		fmt.Println(out)

		fmt.Printf("Run %s to view details, or %s to install.\n\n",
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
