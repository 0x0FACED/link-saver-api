package wrap

import "fmt"

func E(pkg string, msg string, err error) error {
	return fmt.Errorf("%s -> %s: %w", pkg, msg, err)
}
