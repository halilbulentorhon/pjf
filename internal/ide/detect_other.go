//go:build !darwin

package ide

func detectPlatform() []IDE {
	return nil
}
