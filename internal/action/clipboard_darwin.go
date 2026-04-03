//go:build darwin

package action

import (
	"os/exec"
	"strings"
)

type DarwinClipboard struct{}

func NewClipboard() Clipboard {
	return &DarwinClipboard{}
}

func (c *DarwinClipboard) Copy(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
