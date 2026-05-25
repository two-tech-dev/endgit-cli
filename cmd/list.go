/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List installed plugins in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		entries, err := os.ReadDir("plugins")
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn("No plugins directory found. Run 'endgit install <plugin>' first.")
				return
			}
			log.Fatal("Failed to read plugins directory", err)
		}

		type installed struct {
			name    string
			version string
			file    string
		}

		var plugins []installed

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext != ".whl" && ext != ".so" && ext != ".dll" && ext != ".dylib" {
				continue
			}

			base := strings.TrimSuffix(name, ext)
			p := parseInstalledName(base, ext)
			plugins = append(plugins, p)
		}

		if len(plugins) == 0 {
			log.Warn("No plugins installed in ./plugins")
			return
		}

		fmt.Printf("\n  %s\n\n", color.New(color.Bold).Sprint("Installed Plugins"))

		for _, p := range plugins {
			nameStr := color.New(color.FgHiWhite, color.Bold).Sprint(p.name)
			verStr := color.MagentaString("v%s", p.version)
			fmt.Printf("  %s  %s\n", nameStr, verStr)
			fmt.Printf("  %s\n\n", color.HiBlackString(p.file))
		}

		fmt.Printf("  %s plugin(s) installed\n\n", color.CyanString("%d", len(plugins)))
	},
}

func parseInstalledName(base, ext string) struct {
	name    string
	version string
	file    string
} {
	filename := base + ext

	if ext == ".whl" {
		parts := strings.SplitN(base, "-", 3)
		if len(parts) >= 2 {
			name := strings.TrimPrefix(parts[0], "endstone_")
			return struct {
				name    string
				version string
				file    string
			}{name, parts[1], filename}
		}
	}

	parts := strings.Split(base, "-")
	if len(parts) >= 2 {
		version := parts[len(parts)-1]
		name := strings.Join(parts[:len(parts)-1], "-")
		return struct {
			name    string
			version string
			file    string
		}{name, version, filename}
	}

	return struct {
		name    string
		version string
		file    string
	}{base, "unknown", filename}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
