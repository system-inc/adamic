package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

//go:embed observer/notify.c
var notificationSource string

var syscallPattern = regexp.MustCompile(`^([a-zA-Z0-9_]+)\(`)
var quotedPattern = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
var descriptorPattern = regexp.MustCompile(`<(/[^>]*)>`)

func tracedRun(root string, pkg packageInfo, binary, stem string, value *closure, observerPaths ...string) error {
	observer := ""
	if len(observerPaths) > 0 {
		observer = observerPaths[0]
	} else {
		var err error
		observer, err = buildNotificationObserver(filepath.Dir(stem))
		if err != nil {
			return err
		}
	}
	events, err := os.Create(value.Events)
	if err != nil {
		return err
	}
	defer events.Close()
	errors, err := os.Create(stem + ".stderr")
	if err != nil {
		return err
	}
	defer errors.Close()
	// Seccomp notifications are inherited by the whole process tree without
	// ptrace, so LeakSanitizer remains enabled. The decoder also accepts the
	// equivalent pathname/descriptor annotations produced by strace -yy.
	// Command-mode test2json ignores signals before exec, changing native's
	// inherited SIGINT disposition. Convert stdin instead, with no test child.
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	defer reader.Close()
	defer writer.Close()
	converter := exec.Command("go", "tool", "test2json", "-t", "-p", pkg.ImportPath)
	converter.Dir = pkg.Dir
	converter.Stdin = reader
	converter.Stdout = events
	converter.Stderr = errors
	if err := converter.Start(); err != nil {
		return err
	}
	reader.Close()
	args := []string{"-o", stem + ".trace", "--", binary, "-test.v=test2json", "-test.timeout=" + testDeadline}
	command := exec.Command(observer, args...)
	command.Dir = pkg.Dir
	command.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1")
	command.Stdout = writer
	command.Stderr = writer
	loadBefore, _ := os.ReadFile("/proc/loadavg")
	started := time.Now()
	startErr := command.Start()
	writer.Close()
	var runErr error
	if startErr == nil {
		runErr = command.Wait()
	} else {
		runErr = startErr
	}
	conversionErr := converter.Wait()
	seconds := time.Since(started).Seconds()
	loadAfter, _ := os.ReadFile("/proc/loadavg")
	if err := atomicJSON(stem+".timing.json", map[string]any{"started": started.UTC(), "wall_seconds": seconds, "load_before": strings.TrimSpace(string(loadBefore)), "load_after": strings.TrimSpace(string(loadAfter)), "instrument": append([]string{observer}, args...), "uncached": true}); err != nil {
		return err
	}
	if runErr != nil {
		return fmt.Errorf("traced uncached %s: %w; see %s and %s.stderr", pkg.ImportPath, runErr, value.Events, stem)
	}
	if conversionErr != nil {
		return fmt.Errorf("test event conversion: %w", conversionErr)
	}
	traces, err := filepath.Glob(stem + ".trace")
	if err != nil {
		return err
	}
	if len(traces) == 0 {
		return fmt.Errorf("observer produced no process log")
	}
	for _, path := range traces {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 65536), 16<<20)
		for scanner.Scan() {
			collectLine(root, pkg.Dir, scanner.Text(), value)
		}
		err = scanner.Err()
		file.Close()
		if err != nil {
			value.Uncertain = append(value.Uncertain, "trace decode: "+err.Error())
		}
	}
	// The closure is a set of paths, not a list of opens. Hash each member of
	// that union once after the process tree exits; initial and final snapshots
	// still reject changed inputs. No test result or content hash is reused.
	for key := range value.Observed {
		path := key
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		if err := add(root, path, value.Observed); err != nil {
			delete(value.Observed, key)
			value.Uncertain = append(value.Uncertain, "observed input: "+err.Error())
		}
	}
	return events.Sync()
}
func observeLine(root, cwd, line string, value *closure) {
	decodeLine(root, cwd, line, value, func(path string) error { return add(root, path, value.Observed) })
}
func collectLine(root, cwd, line string, value *closure) {
	decodeLine(root, cwd, line, value, func(path string) error {
		if !utf8.ValidString(path) {
			return fmt.Errorf("input pathname is not valid UTF-8")
		}
		value.Observed[inputPath(root, path)] = ""
		return nil
	})
}
func decodeLine(root, cwd, line string, value *closure, addPath func(string) error) {
	if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
		return
	}
	call := syscallPattern.FindStringSubmatch(line)
	if len(call) == 0 {
		value.Uncertain = append(value.Uncertain, "unparsed trace line: "+line)
		return
	}
	if strings.Contains(line, "<unfinished ...>") || strings.Contains(line, "<...") {
		value.Uncertain = append(value.Uncertain, "incomplete trace line")
		return
	}
	paths := []string{}
	accessPath := ""
	if strings.Count(line, "<") != strings.Count(line, ">") {
		value.Uncertain = append(value.Uncertain, "ambiguous descriptor pathname")
	}
	descriptors := descriptorPattern.FindAllStringSubmatch(line, -1)
	for _, descriptor := range descriptors {
		if inside(root, descriptor[1]) && strings.ContainsAny(descriptor[1], "<\\") {
			value.Uncertain = append(value.Uncertain, fmt.Sprintf("escaped descriptor pathname %q", descriptor[1]))
		}
		paths = append(paths, descriptor[1])
	}
	if call[1] != "getdents64" {
		quoted := quotedPattern.FindAllStringIndex(line, -1)
		// Decode the first pathname and the kernel's resolved descriptors. A
		// multi-path operation cannot establish a complete read closure here.
		switch call[1] {
		case "rename", "renameat", "renameat2", "link", "linkat", "symlink", "symlinkat":
			value.Uncertain = append(value.Uncertain, "multi-path file operation")
		}
		if len(quoted) > 0 {
			if strings.HasPrefix(line[quoted[0][1]:], "...") {
				value.Uncertain = append(value.Uncertain, "truncated pathname")
				return
			}
			name, err := strconv.Unquote(line[quoted[0][0]:quoted[0][1]])
			if err != nil {
				value.Uncertain = append(value.Uncertain, "pathname decoding")
				return
			}
			if !filepath.IsAbs(name) {
				prefix := line[:quoted[0][0]]
				dir := descriptorPattern.FindStringSubmatch(prefix)
				if len(dir) > 0 {
					name = filepath.Join(dir[1], name)
				} else {
					// Non-at relative calls lack a cwd annotation. Never guess after chdir.
					value.Uncertain = append(value.Uncertain, "relative pathname without cwd: "+call[1])
					name = filepath.Join(cwd, name)
				}
			}
			paths = append(paths, name)
			accessPath = filepath.Clean(name)
		} else if len(descriptors) == 0 && call[1] != "fchdir" {
			value.Uncertain = append(value.Uncertain, "file syscall without decoded pathname: "+call[1])
		}
	}
	openedPath := ""
	if result := strings.LastIndex(line, ") ="); result >= 0 {
		if decoded := descriptorPattern.FindStringSubmatch(line[result:]); len(decoded) > 0 {
			openedPath = filepath.Clean(decoded[1])
		}
	}
	for _, path := range paths {
		path = filepath.Clean(path)
		if !inside(root, path) {
			continue
		}
		if (path == accessPath || path == openedPath) && (strings.Contains(line, "O_WRONLY") || strings.Contains(line, "O_RDWR") || strings.Contains(line, "O_CREAT") || strings.Contains(line, "O_TRUNC")) {
			value.Uncertain = append(value.Uncertain, "test writes repository input")
		}
		relative, _ := filepath.Rel(root, path)
		gitMetadata := false
		for _, part := range strings.Split(relative, string(filepath.Separator)) {
			if part == ".git" {
				value.Uncertain = append(value.Uncertain, "test reads Git metadata")
				gitMetadata = true
				break
			}
		}
		if gitMetadata {
			continue
		}
		if err := addPath(path); err != nil {
			value.Uncertain = append(value.Uncertain, "observed input: "+err.Error())
		}
		// Missing probes can change their error when a parent changes type,
		// permissions or symlink target, even while the leaf stays missing.
		for parent := filepath.Dir(path); inside(root, parent); parent = filepath.Dir(parent) {
			if err := addPath(parent); err != nil {
				value.Uncertain = append(value.Uncertain, "observed parent: "+err.Error())
			}
			if parent == root {
				break
			}
		}
	}
}

// A sanitizer-using gate cannot produce a green record under an incompatible
// observer. Test that premise before spending a full gate tracing packages.
func observerCompatible(directory, observer string) error {
	source := filepath.Join(directory, "observer-probe.c")
	binary := filepath.Join(directory, "observer-probe")
	if err := os.WriteFile(source, []byte("#include <stdlib.h>\nint main(void) { void *value = malloc(32); free(value); return 0; }\n"), 0600); err != nil {
		return err
	}
	if err := logged(directory, directory+"/observer-build.log", "clang", "-fsanitize=address,undefined", "-g", source, "-o", binary); err != nil {
		return err
	}
	if err := logged(directory, directory+"/observer-plain.log", binary); err != nil {
		return err
	}
	log := filepath.Join(directory, "observer-traced.log")
	if err := logged(directory, log, observer, "-o", filepath.Join(directory, "observer.trace"), "--", binary); err != nil {
		return fmt.Errorf("observer cannot preserve the sanitizer gate; no record published; see %s: %w", log, err)
	}
	return nil
}

func buildNotificationObserver(directory string) (string, error) {
	source := filepath.Join(directory, "notification-observer.c")
	binary := filepath.Join(directory, "notification-observer")
	if err := os.WriteFile(source, []byte(notificationSource), 0600); err != nil {
		return "", err
	}
	if err := logged(directory, filepath.Join(directory, "notification-observer-build.log"), "clang", "-O2", "-Wall", "-Wextra", "-Werror", source, "-o", binary); err != nil {
		return "", err
	}
	return binary, nil
}
