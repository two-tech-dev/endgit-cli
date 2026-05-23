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

var infoResultBox = box.NewBox().
		Style(box.Double).
		Padding(2, 1).
		TitlePosition(box.Top).
		ContentAlign(box.Left).
		Color(box.Cyan).
		TitleColor(box.BrightCyan)

var infoCmd = &cobra.Command{
	Use:   "info <plugin>",
	Short: "Show detailed information about a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Fetching %s...", name)
		s.Start()

		client := api.NewClient()
		p, err := client.GetPlugin(name)

		s.Stop()

		if err != nil {
			log.Fatal("Failed to fetch plugin", err)
		}

		typeBadge := color.HiGreenString("[PY]")
		if strings.ToUpper(p.PluginType) != "PYTHON" {
			typeBadge = color.HiBlueString("[C++]")
		}

		displayName := p.DisplayName
		if displayName == "" {
			displayName = p.Name
		}

		var lines []string

		header := fmt.Sprintf("%s  %s", typeBadge, color.New(color.FgHiWhite, color.Bold).Sprint(displayName))
		if p.LatestVersion != "" {
			header += fmt.Sprintf("  %s", color.MagentaString("v%s", p.LatestVersion))
		}
		lines = append(lines, header)
		lines = append(lines, color.HiBlackString(p.Slug))
		lines = append(lines, "")

		if p.Description != "" {
			lines = append(lines, wrapText(p.Description, 40))
			lines = append(lines, "")
		}

		addField := func(key, value string) {
			if value != "" {
				lines = append(lines, fmt.Sprintf("%-14s %s", color.CyanString(key), value))
			}
		}

		addField("Author", formatAuthor(p.Author))
		addField("License", p.License)
		addField("Type", p.PluginType)
		addField("Status", p.Status)
		addField("Downloads", fmt.Sprintf("%d", p.Downloads))
		addField("Stars", fmt.Sprintf("%d", p.Stars))

		if p.RepoURL != "" {
			addField("Repository", p.RepoURL)
		}

		if len(p.Tags) > 0 {
			addField("Tags", strings.Join(p.Tags, ", "))
		}
		if len(p.Keywords) > 0 {
			addField("Keywords", strings.Join(p.Keywords, ", "))
		}

		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Install: %s", color.HiWhiteString("endgit install %s", p.Name)))

		content := strings.Join(lines, "\n")
		out := infoResultBox.MustRender(displayName, content)
		fmt.Println(out)
	},
}

func wrapText(s string, width int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

func formatAuthor(a api.Author) string {
	if a.DisplayName != "" && a.DisplayName != a.Username {
		return fmt.Sprintf("%s (%s)", a.DisplayName, a.Username)
	}
	if a.Username != "" {
		return a.Username
	}
	return "unknown"
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
