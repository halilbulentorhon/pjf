package action

type Clipboard interface {
	Copy(text string) error
}
