/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	box "github.com/box-cli-maker/box-cli-maker/v3"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/api"
	"github.com/two-tech-dev/endgit-cli/internal/config"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var loginBox = box.NewBox().
		Style(box.Double).
		Padding(2, 1).
		TitlePosition(box.Top).
		ContentAlign(box.Center).
		Color(box.Cyan).
		TitleColor(box.BrightCyan).
		ContentColor(box.BrightWhite)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the EndGit platform",
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient()

		deviceResp, err := client.RequestDeviceCode()
		if err != nil {
			log.Fatal("Failed to initiate device authorization", err)
		}

		out := loginBox.MustRender(
			"Device Authorization",
			fmt.Sprintf("Open in browser:\n%s\n\nEnter code: %s\n\nPress Enter to continue...", deviceResp.VerificationURI, color.New(color.Bold).Sprint(deviceResp.UserCode)),
		)
		fmt.Print(out)

		fmt.Scanln()

		openBrowser(deviceResp.VerificationURI)

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Waiting for authorization..."
		s.Start()

		interval := time.Duration(deviceResp.Interval) * time.Second
		if interval < 5*time.Second {
			interval = 5 * time.Second
		}

		deadline := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)
		var tokenResp *api.DeviceTokenResponse

		for time.Now().Before(deadline) {
			time.Sleep(interval)

			tokenResp, err = client.PollDeviceToken(deviceResp.DeviceCode)
			if err == nil {
				break
			}

			if authErr, ok := err.(*api.DeviceAuthError); ok {
				switch authErr.Code {
				case "authorization_pending":
					continue
				case "slow_down":
					interval += 5 * time.Second
					continue
				case "access_denied":
					s.Stop()
					log.Warn("Authorization was denied")
					return
				case "expired_token":
					s.Stop()
					log.Fatal("Device code expired. Please run 'endgit login' again.", nil)
				}
			}

			s.Stop()
			log.Fatal("Authorization failed", err)
		}

		s.Stop()

		if tokenResp == nil {
			log.Fatal("Authorization timed out. Please run 'endgit login' again.", nil)
		}

		cfg := config.GetConfig()
		cfg.APIToken = tokenResp.AccessToken
		cfg.RefreshToken = tokenResp.RefreshToken
		if err := config.SaveConfig(cfg); err != nil {
			log.Fatal("Failed to save credentials", err)
		}

		if tokenResp.Username != "" {
			log.Successf("Successfully logged in as @%s", tokenResp.Username)
		} else {
			log.Success("Successfully logged in")
		}
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
