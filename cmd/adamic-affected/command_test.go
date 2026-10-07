package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordAndSelectCommand(t *testing.T) {
	// Not parallel: this builds and runs the real command with a controlled Git main.
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("observer compiler unavailable")
	}
	root := t.TempDir()
	writeInput(t, filepath.Join(root, "go.mod"), "module affected-command-proof\n\ngo 1.27\n")
	writeInput(t, filepath.Join(root, "pkg", "probe_test.go"), `package probe
import("os";"testing")
func TestFixture(t *testing.T){b,e:=os.ReadFile("../oracle/value.txt");if e!=nil{t.Fatal(e)};if string(b)!="before"{t.Fatalf("fixture changed: %q",b)}}
func TestListing(t *testing.T){entries,e:=os.ReadDir("../oracle/fixtures");if e!=nil{t.Fatal(e)};if len(entries)!=1{t.Fatalf("fixture enumeration changed: %d",len(entries))}}
`)
	writeInput(t, filepath.Join(root, "quiet", "quiet_test.go"), "package quiet\nimport \"testing\"\nfunc TestQuiet(t *testing.T){}\n")
	fixture := filepath.Join(root, "oracle", "value.txt")
	writeInput(t, fixture, "before")
	writeInput(t, filepath.Join(root, "oracle", "fixtures", "first.a"), "first")
	writeInput(t, filepath.Join(root, "docs", "README.md"), "before")
	writeInput(t, filepath.Join(root, "cohere", "pin.a"), "pin")
	git := func(directory string, args ...string) string {
		t.Helper()
		output, err := command(directory, "git", args...)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(output))
	}
	cohere := filepath.Join(root, "cohere")
	for _, directory := range []string{cohere, root} {
		git(directory, "init", "-q")
		git(directory, "config", "user.name", "Affected proof")
		git(directory, "config", "user.email", "affected-proof@example.invalid")
	}
	git(cohere, "add", "pin.a")
	git(cohere, "commit", "-qm", "Pin fixture")
	writeInput(t, filepath.Join(root, ".gitmodules"), "[submodule \"cohere\"]\n\tpath = cohere\n\turl = ./cohere\n")
	git(root, "add", ".")
	git(root, "commit", "-qm", "Main fixture")
	git(root, "update-ref", "refs/remotes/origin/main", "HEAD")
	binary := filepath.Join(t.TempDir(), "adamic-affected")
	if err := logged(".", binary+".build.log", "go", "build", "-o", binary, "."); err != nil {
		t.Fatal(err)
	}
	recordPath := filepath.Join(t.TempDir(), "main.json")
	executeBinary := func(target string, args ...string) (string, error) {
		cmd := exec.Command(target, args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1")
		output, err := cmd.CombinedOutput()
		return string(output), err
	}
	execute := func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "record" {
			args = append(args, "-isolate", "affected-command-proof/quiet")
		}
		return executeBinary(binary, args...)
	}
	output, err := execute("record", "-out", recordPath)
	if err != nil {
		t.Fatalf("record: %v\n%s", err, output)
	}
	if strings.LastIndex(output, "checkpoint 1/2 affected-command-proof/pkg") > strings.LastIndex(output, "record 2/2 affected-command-proof/quiet") {
		t.Fatal("isolated package started before ordinary evidence was durable")
	}
	if output, err := executeBinary(binary, "record", "-out", recordPath); err == nil || !strings.Contains(output, "checkpoint identity changed") {
		t.Fatalf("resume accepted changed isolation profile: %v %s", err, output)
	}
	selectPackages := func(record string) string {
		t.Helper()
		output, err := execute("select", "-record", record)
		if err != nil {
			t.Fatalf("select: %v\n%s", err, output)
		}
		return output
	}
	if output := selectPackages(recordPath); output != "" {
		t.Fatalf("unchanged fixture main selected packages: %s", output)
	}
	if output, err := execute("record", "-out", recordPath); err == nil || !strings.Contains(output, "record already complete") {
		t.Fatalf("completed reference was silently reused as a fresh run: %v %s", err, output)
	}
	checkpointBytes, err := os.ReadFile(recordPath + ".partial")
	if err != nil {
		t.Fatal(err)
	}
	var checkpoint recordCheckpoint
	if err := json.Unmarshal(checkpointBytes, &checkpoint); err != nil {
		t.Fatal(err)
	}
	readerEvents := checkpoint.Record.Packages["affected-command-proof/pkg"].Events
	quietEvents := checkpoint.Record.Packages["affected-command-proof/quiet"].Events
	quietEvidence, err := os.ReadFile(quietEvents)
	if err != nil {
		t.Fatal(err)
	}
	referenceBytes, err := os.ReadFile(readerEvents)
	if err != nil {
		t.Fatal(err)
	}
	writeInput(t, readerEvents, string(referenceBytes)+"tampered\n")
	if output, err := execute("record", "-out", recordPath); err == nil || !strings.Contains(output, "reference log changed") {
		t.Fatalf("resume accepted tampered evidence: %v %s", err, output)
	}
	writeInput(t, readerEvents, string(referenceBytes))
	checkpoint.Record.Complete = false
	delete(checkpoint.Record.Packages, "affected-command-proof/quiet")
	delete(checkpoint.Record.EventHashes, "affected-command-proof/quiet")
	if err := atomicJSON(recordPath+".partial", checkpoint); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(recordPath); err != nil {
		t.Fatal(err)
	}
	if output := selectPackages(recordPath + ".partial"); !strings.Contains(output, "affected-command-proof/pkg") || !strings.Contains(output, "affected-command-proof/quiet") {
		t.Fatal("incomplete evidence allowed skipping")
	}
	if output, err := execute("record", "-out", recordPath); err != nil || !strings.Contains(output, "resume 1/2") {
		t.Fatalf("resume failed: %v %s", err, output)
	}
	afterResume, err := os.ReadFile(readerEvents)
	if err != nil || string(afterResume) != string(referenceBytes) {
		t.Fatal("resume reran the completed reference package")
	}
	retainedQuiet, err := os.ReadFile(quietEvents)
	if err != nil || string(retainedQuiet) != string(quietEvidence) {
		t.Fatal("resume overwrote the interrupted attempt's unrecorded events")
	}
	t.Log("checkpoint resumed pending package; corrupted event log rejected; incomplete record selected all")
	checkpoint.Failed = map[string]string{"affected-command-proof/quiet": "diagnosed setup failure"}
	if err := atomicJSON(recordPath+".partial", checkpoint); err != nil {
		t.Fatal(err)
	}
	if output, err := execute("record", "-out", recordPath); err == nil || !strings.Contains(output, "reference had failed packages") {
		t.Fatalf("failed reference resumed implicitly: %v %s", err, output)
	}
	if output, err := execute("record", "-retry-failed", "-out", recordPath); err != nil {
		t.Fatalf("explicit retry failed: %v %s", err, output)
	}
	var retried recordFile
	data, err := os.ReadFile(recordPath)
	if err != nil || json.Unmarshal(data, &retried) != nil || len(retried.Retries) != 1 || retried.Retries[0].Failed["affected-command-proof/quiet"] == "" {
		t.Fatal("retry discarded failure evidence")
	}
	if retried.Packages["affected-command-proof/pkg"].Events != readerEvents || retried.Packages["affected-command-proof/quiet"].Events == filepath.Join(checkpoint.Logs, "001.jsonl") {
		t.Fatal("retry replaced previous logs or reran completed evidence")
	}
	t.Log("failed reference required explicit retry; original logs and failure history retained")
	checkpoint.Failed = nil
	writeInput(t, filepath.Join(root, "docs", "README.md"), "after")
	if output := selectPackages(recordPath); output != "" {
		t.Fatalf("unread docs byte selected packages: %s", output)
	}
	writeInput(t, fixture, "after")
	if output := selectPackages(recordPath); !strings.Contains(output, "affected-command-proof/pkg\n") || strings.Contains(output, "affected-command-proof/quiet\n") {
		t.Fatalf("relative input selection: %s", output)
	}
	writeInput(t, fixture, "before")
	contents, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatal(err)
	}
	var recorded recordFile
	if err := json.Unmarshal(contents, &recorded); err != nil {
		t.Fatal(err)
	}
	writeRecord := func(value recordFile) string {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "mutant.json")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	reader := recorded.Packages["affected-command-proof/pkg"]
	// The deadline is part of the invocation identity, even with unchanged files.
	invocation := recorded.Toolchain["test invocation"]
	recorded.Toolchain["test invocation"] = strings.Replace(invocation, "60m", "30m", 1)
	deadlineOutput := selectPackages(writeRecord(recorded))
	if !strings.Contains(deadlineOutput, "affected-command-proof/pkg\n") || !strings.Contains(deadlineOutput, "affected-command-proof/quiet\n") {
		t.Fatalf("deadline identity change failed to select all: %s", deadlineOutput)
	}
	recorded.Toolchain["test invocation"] = invocation
	t.Log("30-minute deadline identity selected all under the 60-minute runner")
	mutatedRecord := recorded
	mutatedRecord.Packages = map[string]closure{}
	for name, value := range recorded.Packages {
		mutatedRecord.Packages[name] = value
	}
	withoutObserved := reader
	withoutObserved.Observed = map[string]string{}
	mutatedRecord.Packages["affected-command-proof/pkg"] = withoutObserved
	writeInput(t, fixture, "after")
	if output := selectPackages(writeRecord(mutatedRecord)); strings.Contains(output, "affected-command-proof/pkg") {
		t.Fatal("observed mutant did not wrongly skip reader")
	}
	proofLog := filepath.Join(t.TempDir(), "observed-proof.log")
	if err := logged(root, proofLog, "go", "test", "-count=1", "./pkg"); err == nil {
		t.Fatal("fresh package gate did not catch observed mutant")
	}
	proofOutput, err := os.ReadFile(proofLog)
	if err != nil || !strings.Contains(string(proofOutput), "fixture changed") {
		t.Fatalf("unexpected proof failure: %v %s", err, proofOutput)
	}
	t.Log("CLI observed-input mutant wrongly skipped reader; fresh uncached package gate caught it")
	writeInput(t, fixture, "before")
	withoutDirectories := reader
	withoutDirectories.Observed = map[string]string{}
	for path, hash := range reader.Observed {
		absolute := path
		if !filepath.IsAbs(path) {
			absolute = filepath.Join(root, path)
		}
		info, err := os.Stat(absolute)
		if err == nil && info.IsDir() {
			continue
		}
		withoutDirectories.Observed[path] = hash
	}
	mutatedRecord.Packages["affected-command-proof/pkg"] = withoutDirectories
	newFixture := filepath.Join(root, "oracle", "fixtures", "new.a")
	writeInput(t, newFixture, "new")
	if output := selectPackages(recordPath); !strings.Contains(output, "affected-command-proof/pkg") {
		t.Fatal("new fixture was not selected")
	}
	if output := selectPackages(writeRecord(mutatedRecord)); strings.Contains(output, "affected-command-proof/pkg") {
		t.Fatal("directory mutant did not wrongly skip reader")
	}
	proofLog = filepath.Join(t.TempDir(), "listing-proof.log")
	if err := logged(root, proofLog, "go", "test", "-count=1", "./pkg"); err == nil {
		t.Fatal("fresh package gate did not catch directory mutant")
	}
	proofOutput, err = os.ReadFile(proofLog)
	if err != nil || !strings.Contains(string(proofOutput), "fixture enumeration changed: 2") {
		t.Fatalf("unexpected proof failure: %v %s", err, proofOutput)
	}
	t.Log("CLI directory-input mutant wrongly skipped new fixture; fresh uncached package gate caught it")
	if err := os.Remove(newFixture); err != nil {
		t.Fatal(err)
	}
	recorded.Toolchain["node --version"] = "v24.mutant\n"
	mutant, err := json.Marshal(recorded)
	if err != nil {
		t.Fatal(err)
	}
	mutatedPath := filepath.Join(t.TempDir(), "node-mutant.json")
	if err := os.WriteFile(mutatedPath, mutant, 0600); err != nil {
		t.Fatal(err)
	}
	output = selectPackages(mutatedPath)
	if !strings.Contains(output, "affected-command-proof/pkg\n") || !strings.Contains(output, "affected-command-proof/quiet\n") {
		t.Fatalf("Node version change failed to select everything: %s", output)
	}
	mutantRoot := t.TempDir()
	writeInput(t, filepath.Join(mutantRoot, "go.mod"), "module affected-selector-mutant\n\ngo 1.27\n")
	for _, name := range []string{"main.go", "trace.go", "record.go", "observer/notify.c"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if name == "main.go" {
			original := "!reflect.DeepEqual(record.Toolchain, tools)"
			if strings.Count(text, original) != 1 {
				t.Fatal("toolchain mutant has no unique mutation point")
			}
			text = strings.Replace(text, original, "!reflect.DeepEqual(tools, tools)", 1)
		}
		writeInput(t, filepath.Join(mutantRoot, name), text)
	}
	mutantBinary := filepath.Join(t.TempDir(), "mutant")
	if err := logged(mutantRoot, mutantBinary+".build.log", "go", "build", "-o", mutantBinary, "."); err != nil {
		t.Fatal(err)
	}
	mutantOutput, err := executeBinary(mutantBinary, "select", "-record", mutatedPath)
	if err != nil || mutantOutput != "" {
		t.Fatalf("toolchain mutant did not reproduce wrongful skip: %v %s", err, mutantOutput)
	}
	actualNode, err := command(root, "node", "--version")
	if err != nil || string(actualNode) == recorded.Toolchain["node --version"] {
		t.Fatal("independent Node identity failed to catch mutant")
	}
	t.Log("compiled toolchain mutant wrongly skipped everything; independent Node version requires both packages")
	t.Log("real CLI: unchanged and docs skipped, relative fixture selected reader, changed Node record selected both packages")
	if output := selectPackages(filepath.Join(t.TempDir(), "missing.json")); !strings.Contains(output, "affected-command-proof/quiet\n") {
		t.Fatal("missing record failed to select all")
	}
	// Resume reuses evidence only within one unfinished reference. Prove its
	// event integrity key matters with a separately compiled validator mutant.
	writeInput(t, filepath.Join(root, "docs", "README.md"), "before")
	if err := atomicJSON(recordPath+".partial", checkpoint); err != nil {
		t.Fatal(err)
	}
	writeInput(t, readerEvents, string(referenceBytes)+`{"Action":"output","Package":"affected-command-proof/pkg","Output":"tampered\n"}`+"\n")
	if output, err := execute("record", "-out", recordPath); err == nil || !strings.Contains(output, "reference log changed") {
		t.Fatalf("event integrity control failed: %v %s", err, output)
	}
	validatorPath := filepath.Join(mutantRoot, "record.go")
	validator, err := os.ReadFile(validatorPath)
	if err != nil {
		t.Fatal(err)
	}
	key := "digest(contents) != record.EventHashes[name]"
	if strings.Count(string(validator), key) != 1 {
		t.Fatal("event integrity mutant target missing")
	}
	writeInput(t, validatorPath, strings.Replace(string(validator), key, "digest(contents) != digest(contents)", 1))
	if err := logged(mutantRoot, mutantBinary+".resume-build.log", "go", "build", "-o", mutantBinary, "."); err != nil {
		t.Fatal(err)
	}
	if output, err := executeBinary(mutantBinary, "record", "-out", recordPath, "-isolate", "affected-command-proof/quiet"); err != nil {
		t.Fatalf("integrity mutant did not accept corrupted reference: %v %s", err, output)
	}
	corrupted, err := os.ReadFile(readerEvents)
	if err != nil || digest(corrupted) == checkpoint.Record.EventHashes["affected-command-proof/pkg"] {
		t.Fatal("independent event hash failed to catch resume mutant")
	}
	t.Log("compiled resume mutant dropped event hash and accepted corrupted evidence; independent reference hash caught it")
	// Removing attempt-directory rollover overwrites evidence whose process
	// ended before the recorder could mark it completed or failed.
	rollover := `checkpoint.Logs, err = os.MkdirTemp(destination+".logs", "resume-")`
	validator, err = os.ReadFile(validatorPath)
	if err != nil || strings.Count(string(validator), rollover) != 1 {
		t.Fatal("attempt rollover mutation point missing")
	}
	writeInput(t, validatorPath, strings.Replace(string(validator), rollover, "checkpoint.Logs, err = checkpoint.Logs, nil", 1))
	if err := logged(mutantRoot, mutantBinary+".rollover-build.log", "go", "build", "-o", mutantBinary, "."); err != nil {
		t.Fatal(err)
	}
	writeInput(t, readerEvents, string(referenceBytes))
	if err := atomicJSON(recordPath+".partial", checkpoint); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(recordPath); err != nil {
		t.Fatal(err)
	}
	if output, err := executeBinary(mutantBinary, "record", "-out", recordPath, "-isolate", "affected-command-proof/quiet"); err != nil {
		t.Fatalf("rollover mutant failed unexpectedly: %v %s", err, output)
	}
	overwritten, err := os.ReadFile(quietEvents)
	if err != nil || digest(overwritten) == digest(quietEvidence) {
		t.Fatal("independent digest did not catch overwritten attempt evidence")
	}
	t.Log("compiled rollover mutant overwrote unrecorded attempt events; original evidence digest caught it")
}
