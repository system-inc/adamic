// Package coherepin checks capture provenance against the repository's gitlink.
package coherepin

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func Pinned(root string) (string, error) {
	data, err := exec.Command("git", "-C", root, "ls-tree", "HEAD", "cohere").Output()
	if err != nil {
		return "", fmt.Errorf("read cohere gitlink: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) != 4 || fields[0] != "160000" || fields[1] != "commit" || fields[3] != "cohere" || len(fields[2]) != 40 {
		return "", fmt.Errorf("invalid cohere gitlink: %q", data)
	}
	return fields[2], nil
}

// Check requires both the recorded capture and the checked-out source to match HEAD.
func Check(root, captured string) error {
	pin, err := Pinned(root)
	if err != nil {
		return err
	}
	if captured != pin {
		return fmt.Errorf("capture pin %s differs from cohere gitlink %s; recapture from Go", captured, pin)
	}
	data, err := exec.Command("git", "-C", filepath.Join(root, "cohere"), "rev-parse", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("read cohere checkout: %w", err)
	}
	if got := strings.TrimSpace(string(data)); got != pin {
		return fmt.Errorf("cohere checkout %s differs from gitlink %s", got, pin)
	}
	return nil
}
