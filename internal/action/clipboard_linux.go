//go:build linux

package action

import "fmt"

type LinuxClipboard struct{}

func NewClipboard() Clipboard {
	return &LinuxClipboard{}
}

func (c *LinuxClipboard) Copy(text string) error {
	return fmt.Errorf("clipboard not yet supported on Linux")
}
