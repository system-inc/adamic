package buildcache

import (
	"regexp"
	"strings"
	"testing"
)

// The expressions Settle parsed strace's lines with until its parse was the cost of a trace (#vt46geg step 3): each
// hand parser in strace.go must take exactly what its expression takes.
var (
	traceLineExpression   = regexp.MustCompile(`^(\d+)\s+(.*)$`)
	resumedLineExpression = regexp.MustCompile(`^<\.\.\. (\w+) resumed>\s?(.*)$`)
	callLineExpression    = regexp.MustCompile(`^(\w+)\((.*)\)\s+=\s+(-?\d+|\?)(?:<([^>]*)>)?(?:\s+(E[A-Z0-9]+))?`)
	exitLineExpression    = regexp.MustCompile(`^\+\+\+ (?:exited with \d+|killed by \w+(?: \(core dumped\))?) \+\+\+$`)
	directoryExpression   = regexp.MustCompile(`^(AT_FDCWD|-?\d+)(?:<([^>]*)>)?,\s*`)
	quotedExpression      = regexp.MustCompile(`^"((?:[^"\\]|\\.)*)"`)
	descriptorExpression  = regexp.MustCompile(`^-?\d+<([^>]*)>`)
)

// parsersAgree holds every parser against its expression on text, and on the parts of text a parser would hand on:
// the rest after a pid, a call's arguments, and what follows each comma.
func parsersAgree(t *testing.T, text string) {
	t.Helper()
	pieces := []string{text}
	if match := traceLineExpression.FindStringSubmatch(text); match != nil {
		pieces = append(pieces, match[2])
	}
	for _, piece := range append([]string{}, pieces...) {
		if match := callLineExpression.FindStringSubmatch(piece); match != nil {
			pieces = append(pieces, match[2])
		}
	}
	for _, piece := range append([]string{}, pieces...) {
		for index := range len(piece) {
			if piece[index] == ',' || piece[index] == '(' || piece[index] == '<' {
				pieces = append(pieces, piece[index+1:], strings.TrimLeft(piece[index+1:], " "))
			}
		}
	}
	for _, piece := range pieces {
		if pid, rest, ok := splitTraceLine(piece); ok != traceLineExpression.MatchString(piece) {
			t.Fatalf("splitTraceLine(%q) = %v, the expression %v", piece, ok, !ok)
		} else if match := traceLineExpression.FindStringSubmatch(piece); ok && (pid != match[1] || rest != match[2]) {
			t.Fatalf("splitTraceLine(%q) = %q, %q; the expression %q", piece, pid, rest, match[1:])
		}
		if got, want := isExit(piece), exitLineExpression.MatchString(piece); got != want {
			t.Fatalf("isExit(%q) = %v, the expression %v", piece, got, want)
		}
		if tail, ok := resumedTail(piece); ok != resumedLineExpression.MatchString(piece) {
			t.Fatalf("resumedTail(%q) = %v, the expression %v", piece, ok, !ok)
		} else if match := resumedLineExpression.FindStringSubmatch(piece); ok && tail != match[2] {
			t.Fatalf("resumedTail(%q) = %q, the expression %q", piece, tail, match[2])
		}
		call, ok := parseCall(piece)
		if match := callLineExpression.FindStringSubmatch(piece); ok != (match != nil) {
			t.Fatalf("parseCall(%q) = %v, the expression %v", piece, ok, !ok)
		} else if got := []string{call.call, call.arguments, call.value, call.opened, call.errno}; ok && strings.Join(got, "\x00") != strings.Join(match[1:], "\x00") {
			t.Fatalf("parseCall(%q) = %q, the expression %q", piece, got, match[1:])
		}
		decoded, length, ok := directoryArgument(piece)
		if match := directoryExpression.FindStringSubmatch(piece); ok != (match != nil) {
			t.Fatalf("directoryArgument(%q) = %v, the expression %v", piece, ok, !ok)
		} else if ok && (decoded != match[2] || length != len(match[0])) {
			t.Fatalf("directoryArgument(%q) = %q, %d; the expression %q, %d", piece, decoded, length, match[2], len(match[0]))
		}
		inside, length, ok := quotedArgument(piece)
		if match := quotedExpression.FindStringSubmatch(piece); ok != (match != nil) {
			t.Fatalf("quotedArgument(%q) = %v, the expression %v", piece, ok, !ok)
		} else if ok && (inside != match[1] || length != len(match[0])) {
			t.Fatalf("quotedArgument(%q) = %q, %d; the expression %q, %d", piece, inside, length, match[1], len(match[0]))
		}
		path, ok := descriptorPath(piece)
		if match := descriptorExpression.FindStringSubmatch(piece); ok != (match != nil) {
			t.Fatalf("descriptorPath(%q) = %v, the expression %v", piece, ok, !ok)
		} else if ok && path != match[1] {
			t.Fatalf("descriptorPath(%q) = %q, the expression %q", piece, path, match[1])
		}
	}
}

// Lines strace 6.8 prints under cmd/traced's flags, and the near misses around each part of them.
var straceLines = []string{
	`3315478 openat(AT_FDCWD</tmp/a>, "w\303\251.txt", O_RDONLY) = 3</tmp/a/w\303\251.txt>`,
	`3315478 openat(AT_FDCWD</tmp/a>, "a>b\"c<d.txt", O_RDONLY) = 3</tmp/a/a\76b\"c\74d.txt>`,
	`3315479 getdents64(3</tmp/a/d\303\251>, 0x65530c0ca910 /* 2 entries */, 32768) = 48`,
	`12 newfstatat(AT_FDCWD, "/x", 0x7ffd5e1c, 0) = -1 ENOENT (No such file or directory)`,
	`12 newfstatat(AT_FDCWD, "/x", 0x7ffd5e1c, 0) = -1 ENOTDIR (Not a directory)`,
	`12 faccessat2(AT_FDCWD, "/x", X_OK, AT_EACCESS) = -1 EACCES (Permission denied)`,
	`12 openat(-100, "/x", O_RDONLY) = 3`,
	`12 openat(4</a/b>,  "c", O_RDONLY) = 5</a/b/c>`,
	`12 openat(4</a/b>, "c", O_RDONLY) = 5</a/b/c (deleted)>`,
	`12 openat(4, "c", O_RDONLY) = 5`,
	`12 openat(4<unclosed, "c", O_RDONLY) = 5`,
	`12 openat(4</a>x, "c", O_RDONLY) = 5`,
	`12 execve("/bin/sh", ["sh", "-c", "a) = 1"], 0x7ffd /* 40 vars */) = 0`,
	`12 execve("/bin/sh", ["sh"], 0x7ffd /* 40 vars */) = ?`,
	`12 openat(AT_FDCWD, "/odd) = 3<x>", O_RDONLY) = -1 ENOENT (No such file or directory)`,
	`12 openat(AT_FDCWD, "/x", O_RDONLY <unfinished ...>`,
	`12 <... openat resumed>) = 3</x>`,
	`12 <... openat resumed>, 0) = 3</x>`,
	`12 <... openat resumed> ) = 3`,
	`12 <...  openat resumed>) = 3`,
	`12 <... resumed>) = 3`,
	`12 clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, child_tidptr=0x7f) = 3315480`,
	`12 clone3({flags=CLONE_VM|CLONE_VFORK, exit_signal=SIGCHLD, stack=0x7f, stack_size=0x9000}, 88) = 13`,
	`12 fchdir(3</home/ahra>) = 0`,
	`12 fchdir(-1</nowhere>) = -1 EBADF (Bad file descriptor)`,
	`12 rename("/a", "/b") = 0`,
	`12 renameat2(AT_FDCWD</a>, "x", 4</b>, "y", RENAME_NOREPLACE) = 0`,
	`12 readlinkat(AT_FDCWD, "/proc/self/exe", "/usr/bin/go", 128) = 11`,
	`12 +++ exited with 0 +++`,
	`12 +++ exited with 137 +++`,
	`12 +++ killed by SIGKILL +++`,
	`12 +++ killed by SIGSEGV (core dumped) +++`,
	`12 +++ killed by SIGSEGV (core dumped) (core dumped) +++`,
	`12 +++ exited with  +++`,
	`12 +++ exited with 1 +++ +++`,
	`12 --- SIGCHLD {si_signo=SIGCHLD} ---`,
	"12\topenat(AT_FDCWD, \"/x\", O_RDONLY)\t=\t3",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = 3 \tEAGAIN",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = 3 E",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = 3ENOENT",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = 3 Eagain",
	"12 openat(AT_FDCWD, \"/x\\\", O_RDONLY) = 3",
	"12 openat(AT_FDCWD, \"/x\\\n\", O_RDONLY) = 3",
	"12 openat(AT_FDCWD, \"/x\ny\", O_RDONLY) = 3",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = -",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = --1",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) =3",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY)= 3",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = 3<x",
	"12 openat(AT_FDCWD, \"/x\", O_RDONLY) = 3<x\n>",
	"12 openat(AT_FDCWD, \"\xff\\\xff\", O_RDONLY) = 3<\xff>",
	"12\n openat(AT_FDCWD, \"/x\", O_RDONLY) = 3",
	"12 open\n(\"/x\") = 3",
	"openat(AT_FDCWD, \"/x\", O_RDONLY) = 3",
	"12",
	"12 ",
	"",
	"AT_FDCWD",
	"AT_FDCWD<",
	"-",
	"-<x>,",
	"\"",
	"\"\\",
}

// Every hand parser takes exactly what its expression takes: on the lines strace prints, and on the near misses around
// each of their parts. Mutants: any parser's rule loosened or tightened by one case (a sign, a space, a missing close).
func TestTheStraceParsersMatchTheirExpressions(t *testing.T) {
	t.Parallel()
	for _, line := range straceLines {
		parsersAgree(t, line)
	}
	for _, line := range strings.Split(syntheticTrace("/home/ahra/levers/adamic", 200), "\n") {
		parsersAgree(t, line)
	}
}

// FuzzTheStraceParsers holds the hand parsers against their expressions on anything: go test -fuzz FuzzTheStraceParsers.
func FuzzTheStraceParsers(f *testing.F) {
	for _, line := range straceLines {
		f.Add(line)
	}
	f.Fuzz(func(t *testing.T, text string) {
		parsersAgree(t, text)
	})
}
