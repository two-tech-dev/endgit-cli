/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a plugin to the EndGit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Work in progress: Publishing plugins is not yet implemented.")
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
