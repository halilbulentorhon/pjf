//go:build darwin

package action

import (
	"os"
	"os/exec"
)

type DarwinTerminalOpener struct{}

func NewTerminalOpener() TerminalOpener {
	return &DarwinTerminalOpener{}
}

func (t *DarwinTerminalOpener) Open(dir string) error {
	if isITermInstalled() {
		return openITerm(dir)
	}
	return exec.Command("open", "-a", "Terminal", dir).Run()
}

func isITermInstalled() bool {
	_, err := os.Stat("/Applications/iTerm.app")
	return err == nil
}

func openITerm(dir string) error {
	script := `
		on run argv
			set targetDir to item 1 of argv
			tell application "iTerm"
				activate
				tell current window
					create tab with default profile
					tell current session
						write text "cd " & quoted form of targetDir
					end tell
				end tell
			end tell
		end run
	`
	return exec.Command("osascript", "-e", script, dir).Run()
}

func (t *DarwinTerminalOpener) OpenWithCommand(dir string, command string) error {
	if isITermInstalled() {
		return openITermWithCommand(dir, command)
	}
	return openTerminalWithCommand(dir, command)
}

func openITermWithCommand(dir string, command string) error {
	script := `
		on run argv
			set targetDir to item 1 of argv
			set targetCmd to item 2 of argv
			tell application "iTerm"
				activate
				tell current window
					create tab with default profile
					tell current session
						write text "cd " & quoted form of targetDir & " && " & targetCmd
					end tell
				end tell
			end tell
		end run
	`
	return exec.Command("osascript", "-e", script, dir, command).Run()
}

func openTerminalWithCommand(dir string, command string) error {
	script := `
		on run argv
			set targetDir to item 1 of argv
			set targetCmd to item 2 of argv
			tell application "Terminal"
				activate
				do script "cd " & quoted form of targetDir & " && " & targetCmd
			end tell
		end run
	`
	return exec.Command("osascript", "-e", script, dir, command).Run()
}
