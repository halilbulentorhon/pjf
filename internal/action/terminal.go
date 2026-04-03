package action

type TerminalOpener interface {
	Open(dir string) error
}
