package oracle

import (
	"os"
	"os/exec"
	"testing"
)

// The release-flags lane includes the actual profile-built stage 1 binaries.
// Ordinary runs never enable this lane or change their fixture flags.
func TestReleaseStage1ProfilesAgree(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_RELEASE") != "1" {
		t.Skip("set ADAMIC_ORACLE_RELEASE=1")
	}
	if os.Getenv("ADAMIC_STAGE1_PARSE_BINARY") == "" || os.Getenv("ADAMIC_STAGE1_LINT_BINARY") == "" {
		t.Fatal("release oracle requires ADAMIC_STAGE1_PARSE_BINARY and ADAMIC_STAGE1_LINT_BINARY")
	}
	log, err := os.CreateTemp(t.TempDir(), "stage1-profile-oracle-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	command := exec.Command("go", "test", "-v", "-count=1", "-timeout", "30m", "./stage1/cohere/parse", "./stage1/cohere/lint", "-run", "^TestShippedProfileAgreesWithGo$")
	command.Dir = repository
	command.Stdout = log
	command.Stderr = log
	if err = command.Run(); err != nil {
		data, _ := os.ReadFile(log.Name())
		t.Fatalf("stage 1 shipping oracle: %v\n%s", err, data)
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
