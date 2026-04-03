package action

type TerminalOpener interface {
	Open(dir string) error
	OpenWithCommand(dir string, command string) error
}
