package oracle

import (
	"path/filepath"
	"testing"
)

func TestStep21WritingCallTypeScriptWitness(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_builtin_narrow_terminal.ts"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 || string(source.stdout) != "Error\nfinally\n" || len(source.stderr) != 0 {
		t.Fatalf("source Node witness: %+v", source)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := nativelyUncached(t, program)
	want := "adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n"
	for name, got := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != want {
			t.Fatalf("%s checked writing-call stop: exit %d stdout %q stderr %q; want exit 70 and %q", name, got.exitCode, got.stdout, got.stderr, want)
		}
	}
}

func TestStep21PrimitiveWritingCallRefusal(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_soundness_terminal.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	if err == nil {
		t.Fatal("admitted primitive writing-call fixture")
	}
	// Pin the general rule's existing primitive negative too, rather than leave it
	// registered as an admitted .a checked witness after the new ruling.
	want := path + ":6:63: Adamic 0.1 refuses a narrowed read of value after change() can write it; narrow again after the call"
	if err.Error() != want {
		t.Fatalf("primitive writing-call refusal: got %v, want %s", err, want)
	}
}

func TestStep21NonWritingControlAgreement(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/step21_builtin_narrow_control.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 || string(source.stdout) != "TypeError\nfinally\n" || len(source.stderr) != 0 {
		t.Fatalf("source control: %+v", source)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": sanitized, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(source, got); difference != "" {
			t.Fatalf("%s control: %s", name, difference)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatalf("control leaks: %s", leaked)
	}
}
