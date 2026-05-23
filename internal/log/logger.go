/*
Copyright © 2026 Two Tech Studio
*/
package log

import (
	"fmt"
	"os"

	box "github.com/box-cli-maker/box-cli-maker/v3"
	"github.com/fatih/color"
)

var (
	errBox = box.NewBox().
		Style(box.Bold).
		Padding(1, 0).
		TitlePosition(box.Inside).
		Color(box.Red).
		TitleColor(box.BrightRed).
		ContentColor(box.BrightWhite)

	warnBox = box.NewBox().
		Style(box.Single).
		Padding(1, 0).
		TitlePosition(box.Inside).
		Color(box.Yellow).
		TitleColor(box.BrightYellow).
		ContentColor(box.BrightWhite)

	successBox = box.NewBox().
			Style(box.Single).
			Padding(1, 0).
			TitlePosition(box.Inside).
			Color(box.Green).
			TitleColor(box.BrightGreen).
			ContentColor(box.BrightWhite)

	infoBox = box.NewBox().
		Style(box.Single).
		Padding(1, 0).
		TitlePosition(box.Inside).
		Color(box.Cyan).
		TitleColor(box.BrightCyan).
		ContentColor(box.BrightWhite)
)

// Error logs an error message in a box and returns the error for potential exit handling
func Error(msg string, err error) error {
	content := msg
	if err != nil {
		content = fmt.Sprintf("%s\n%v", msg, err)
	}
	out := errBox.MustRender("Error", content)
	fmt.Fprintln(os.Stderr, out)
	return fmt.Errorf("%s", msg)
}

// Errorf logs a formatted error message in a box
func Errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	out := errBox.MustRender("Error", msg)
	fmt.Fprintln(os.Stderr, out)
	return fmt.Errorf("%s", msg)
}

// Warn logs a warning message in a box
func Warn(msg string) {
	out := warnBox.MustRender("Warning", msg)
	fmt.Fprintln(os.Stderr, out)
}

// Warnf logs a formatted warning message in a box
func Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	out := warnBox.MustRender("Warning", msg)
	fmt.Fprintln(os.Stderr, out)
}

// Info logs an info message with a cyan arrow
func Info(msg string) {
	fmt.Printf("%s %s\n", color.CyanString("→"), msg)
}

// Infof logs a formatted info message with a cyan arrow
func Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", color.CyanString("→"), msg)
}

// Success logs a success message with a green checkmark
func Success(msg string) {
	fmt.Printf("%s %s\n", color.GreenString("✓"), msg)
}

// Successf logs a formatted success message with a green checkmark
func Successf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", color.GreenString("✓"), msg)
}

// SuccessBox displays a success message in a green box
func SuccessBox(title, msg string) {
	out := successBox.MustRender(title, msg)
	fmt.Println(out)
}

// InfoBox displays an info message in a cyan box
func InfoBox(title, msg string) {
	out := infoBox.MustRender(title, msg)
	fmt.Println(out)
}

// Debug logs a debug message (only if DEBUG env var is set)
func Debug(msg string) {
	if os.Getenv("DEBUG") != "" {
		fmt.Printf("%s %s\n", color.HiBlackString("[DEBUG]"), msg)
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		msg := fmt.Sprintf(format, args...)
		fmt.Printf("%s %s\n", color.HiBlackString("[DEBUG]"), msg)
	}
}

// Fatal logs an error in a box and exits with code 1
func Fatal(msg string, err error) {
	Error(msg, err)
	os.Exit(1)
}

// Fatalf logs a formatted error in a box and exits with code 1
func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	os.Exit(1)
}
