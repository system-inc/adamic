package comments

import (
	"testing"
)

// The guard must not change which comments are found, only how fast the scan gets there.
//
// This compares the actual ranges rather than a count, deliberately. A guard that dropped one
// trailing comment per file and gained one elsewhere would hold any count stable, and a rule that
// stops receiving a comment reports nothing for it while every fixture still passes. The comparison
// has to be of the thing itself.
func TestCommentGuardFindsTheSameComments(t *testing.T) {
	t.Parallel()
	files := realSourceFiles(t, 1200)

	for _, file := range files {
		guarded := All(file.sourceFile)
		reference := AllWithoutGuard(file.sourceFile)

		if len(guarded) != len(reference) {
			t.Fatalf("%s: guard found %d comments, ungated scan found %d",
				file.name, len(guarded), len(reference))
		}
		for index := range guarded {
			if guarded[index].Range != reference[index].Range {
				t.Fatalf("%s: comment %d differs, guard %v against ungated %v",
					file.name, index, guarded[index].Range, reference[index].Range)
			}
			if guarded[index].Text != reference[index].Text {
				t.Fatalf("%s: comment %d text differs", file.name, index)
			}
		}
	}
	t.Logf("guard preserves every comment range across %d real files", len(files))
}

// The cases a corpus cannot be relied on to contain.
//
// A differential test over 800 files recently passed a gate that would have broken `s\tc`, because
// no file happened to hold the multi-space form. Absence from a corpus is not absence in general, so
// the separators and the near-miss shapes are written out by hand as well as measured.
func TestCommentGuardHandlesEverySeparator(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{"separated by a space", "const a = 1; // trailing\n", 1},
		{"separated by a tab", "const a = 1;\t// trailing\n", 1},
		{"separated by a newline", "const a = 1;\n// leading\nconst b = 2;\n", 1},
		{"separated by a carriage return and newline", "const a = 1;\r\n// leading\r\nconst b = 2;\r\n", 1},
		{"no separator at all", "const a = 1;// trailing\n", 1},
		{"block comment with no separator", "const a = 1;/* trailing */\n", 1},
		{"comment at position zero", "// first line\nconst a = 1;\n", 1},
		{"two comments on one line", "const a = 1; /* one */ /* two */\n", 2},
		{"no comments at all", "const a = 1;\nconst b = 2;\n", 0},

		// A shebang is trivia the comment scanner walks past, and it is the one non-whitespace thing
		// a comment can sit behind. The corpus caught this and the hand-written cases did not: a
		// real executable script lost its module comment and its first import comment, because the
		// guard stopped at the `#`. Pinned here so it cannot regress quietly.
		{"comment after a shebang", "#!/usr/bin/env -S pnpm tsx\n\n/* module comment */\nconst a = 1;\n", 1},
		{"line comment after a shebang", "#!/usr/bin/env node\n// leading\nconst a = 1;\n", 1},

		// The guard admits these and the scanner then declines them, which is the safe direction and
		// is on purpose: a slash that is not a comment costs one scan, and a comment the guard hid
		// would silence every comment rule for it.
		{"a division is not a comment", "const a = 6 / 2;\n", 0},
		{"a regex literal is not a comment", "const a = /abc/.test(b);\n", 0},
		{"a regex containing a slash", "const a = /a\\/b/.test(c);\n", 0},
		{"a string containing two slashes", "const a = 'http://example.com';\n", 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceFile := parseSourceForTest(t, testCase.sourceText)

			guarded := All(sourceFile)
			reference := AllWithoutGuard(sourceFile)

			if len(guarded) != testCase.wantCount {
				t.Fatalf("expected %d comments, guard found %d", testCase.wantCount, len(guarded))
			}
			if len(guarded) != len(reference) {
				t.Fatalf("guard found %d, ungated scan found %d", len(guarded), len(reference))
			}
			for index := range guarded {
				if guarded[index].Range != reference[index].Range {
					t.Fatalf("comment %d differs from the ungated scan", index)
				}
			}
		})
	}
}

// A guard that never skips is pure overhead, and a guard that skips everything is a bug that the
// differential test alone would not catch on a corpus with few comments.
func TestCommentGuardActuallySkips(t *testing.T) {
	t.Parallel()
	// A run of whitespace with no slash cannot begin a comment.
	if canBeginAt("const a = 1;    const b = 2;", 12) {
		t.Fatalf("guard admits a position with no slash after it")
	}
	// A slash after whitespace must be admitted, whatever the whitespace is.
	for _, text := range []string{"a; // c", "a;\t// c", "a;\n// c", "a;\r\n// c", "a;// c", "a;/* c */"} {
		if !canBeginAt(text, 2) {
			t.Fatalf("guard rejects %q, where a comment does begin", text)
		}
	}
}
