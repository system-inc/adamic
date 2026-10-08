package helpers

import (
	"path/filepath"
	"testing"
)

func TestAccessedName(t *testing.T) {
	root, _ := filepath.Abs("../../../..")
	t.Log(string(run(t, root, "python3", "stage1/cohere/lint/helpers/wave15/accessed_property/validate.py")))
}
