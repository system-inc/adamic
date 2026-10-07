package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var syscallPattern = regexp.MustCompile(`^([a-zA-Z0-9_]+)\(`)
var quotedPattern = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
var descriptorPattern = regexp.MustCompile(`<(/[^>]*)>`)

func tracedRun(root string, pkg packageInfo, binary, stem string, value *closure) error {
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
	// -yy gives absolute cwd/dirfd annotations, including after fork and chdir.
	// -ff prevents interleaved unfinished syscall records. %file includes failed
	// existence checks and modern stat calls; getdents64 includes actual listings.
	args := []string{"-ff", "-yy", "-s", "0", "-e", "trace=%file,getdents64", "-o", stem + ".trace", "go", "tool", "test2json", "-t", "-p", pkg.ImportPath, binary, "-test.v=test2json", "-test.timeout=30m"}
	command := exec.Command("strace", args...)
	command.Dir = pkg.Dir
	command.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1")
	command.Stdout = events
	command.Stderr = errors
	if err := command.Run(); err != nil {
		return fmt.Errorf("traced uncached %s: %w; see %s and %s.stderr (sanitizers may reject ptrace)", pkg.ImportPath, err, value.Events, stem)
	}
	traces, err := filepath.Glob(stem + ".trace.*")
	if err != nil {
		return err
	}
	if len(traces) == 0 {
		return fmt.Errorf("strace produced no process logs")
	}
	for _, path := range traces {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 65536), 16<<20)
		for scanner.Scan() {
			observeLine(root, pkg.Dir, scanner.Text(), value)
		}
		err = scanner.Err()
		file.Close()
		if err != nil {
			value.Uncertain = append(value.Uncertain, "trace decode: "+err.Error())
		}
	}
	return nil
}
func observeLine(root, cwd, line string, value *closure) {
	if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
		return
	}
	call := syscallPattern.FindStringSubmatch(line)
	if len(call) == 0 {
		value.Uncertain = append(value.Uncertain, "unparsed trace line")
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
		if err := add(root, path, value.Observed); err != nil {
			value.Uncertain = append(value.Uncertain, "observed input: "+err.Error())
		}
		// Missing probes can change their error when a parent changes type,
		// permissions or symlink target, even while the leaf stays missing.
		for parent := filepath.Dir(path); inside(root, parent); parent = filepath.Dir(parent) {
			if err := add(root, parent, value.Observed); err != nil {
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
func observerCompatible(directory string) error {
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
	if err := logged(directory, log, "strace", "-f", "-yy", "-s", "0", "-e", "trace=%file,getdents64", "-o", filepath.Join(directory, "observer.trace"), binary); err != nil {
		return fmt.Errorf("observer cannot preserve the sanitizer gate; no record published; see %s: %w", log, err)
	}
	return nil
}
