//go:build linux

package action

import "fmt"

type LinuxTerminalOpener struct{}

func NewTerminalOpener() TerminalOpener {
	return &LinuxTerminalOpener{}
}

func (t *LinuxTerminalOpener) Open(dir string) error {
	return fmt.Errorf("terminal opener not yet supported on Linux")
}

func (t *LinuxTerminalOpener) OpenWithCommand(dir string, command string) error {
	return fmt.Errorf("terminal opener not yet supported on Linux")
}
