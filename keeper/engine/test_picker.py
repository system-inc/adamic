"""test_picker.py: picker.py against a fixture repository with a known landed range (#ysavcpk).

	python3 -I test_picker.py
"""
import json, os, subprocess, sys, tempfile, unittest

here = os.path.dirname(os.path.abspath(__file__))

base_go = """package pkg

func Clamp(value, limit int) int {
	return value
}
"""

head_go = """package pkg

func Clamp(value, limit int) int {
	if value > limit {
		return limit
	}
	total := 0
	total = add(total, value)
	message := "if a < b && c == 7"
	_ = message
	return total + 3
}

func add(a, b int) int { return a + b }
"""

test_go = """package pkg

import "testing"

func TestClamp(t *testing.T) {
	if Clamp(5, 3) != 3 {
		t.Fatal("wrong")
	}
}
"""


def run(*command, cwd=None):
    return subprocess.run(command, cwd=cwd, check=True, capture_output=True, text=True).stdout


class Picker(unittest.TestCase):
    def setUp(self):
        self.repo = tempfile.mkdtemp()
        run("git", "init", "-q", cwd=self.repo)
        run("git", "config", "user.email", "keeper@example.com", cwd=self.repo)
        run("git", "config", "user.name", "keeper", cwd=self.repo)
        self.write("pkg/clamp.go", base_go)
        self.write("pkg/clamp_test.go", test_go)
        self.write("untested/free.go", base_go.replace("package pkg", "package untested"))
        self.commit("base")
        self.base = run("git", "rev-parse", "HEAD", cwd=self.repo).strip()
        self.write("pkg/clamp.go", head_go)
        self.write("pkg/clamp_test.go", test_go.replace("5, 3", "9, 3"))  # a test change: never a mutant source
        self.write("untested/free.go", head_go.replace("package pkg", "package untested"))  # no tests there
        self.commit("head")
        self.head = run("git", "rev-parse", "HEAD", cwd=self.repo).strip()

    def write(self, path, text):
        os.makedirs(os.path.join(self.repo, os.path.dirname(path)), exist_ok=True)
        open(os.path.join(self.repo, path), "w").write(text)

    def commit(self, message):
        run("git", "add", "-A", cwd=self.repo)
        run("git", "commit", "-q", "-m", message, cwd=self.repo)

    def pick(self, base, head, maximum=8):
        out = tempfile.mkdtemp()
        run(sys.executable, "-I", os.path.join(here, "picker.py"), self.repo, base, head, out, "--max", str(maximum))
        return out, json.load(open(os.path.join(out, "manifest.json")))

    def test_mutants_come_only_from_changed_code_under_test(self):
        out, manifest = self.pick(self.base, self.head, maximum=50)
        got = {(mutant["file"], mutant["line"], mutant["op"], mutant["after"]) for mutant in manifest["mutants"]}
        self.assertEqual(got, {
            ("pkg/clamp.go", 4, "flip-condition", "if value <= limit {"),
            ("pkg/clamp.go", 4, "off-by-one", "if value >= limit {"),
            ("pkg/clamp.go", 7, "change-constant", "total := 1"),
            ("pkg/clamp.go", 8, "drop-statement", ""),
            ("pkg/clamp.go", 8, "swap-arguments", "total = add(value, total)"),
            ("pkg/clamp.go", 10, "drop-statement", ""),
            ("pkg/clamp.go", 11, "change-constant", "return total + 4"),
        })  # never the test file, never the untested package, never line 9's string, never unchanged line 14

    def test_the_cap_spreads_across_files(self):
        self.write("pkg/limit.go", "package pkg\n\nfunc Limit(n int) bool {\n\treturn n == 64\n}\n")
        self.commit("second file")
        head = run("git", "rev-parse", "HEAD", cwd=self.repo).strip()
        out, manifest = self.pick(self.base, head, maximum=2)
        self.assertEqual(sorted(mutant["file"] for mutant in manifest["mutants"]), ["pkg/clamp.go", "pkg/limit.go"])

    def test_never_inside_a_string(self):
        out, manifest = self.pick(self.base, self.head, maximum=50)
        for mutant in manifest["mutants"]:
            self.assertNotEqual(mutant["line"], 9, mutant)  # the message literal holds <, &&, == and 7, all decoys

    def test_every_diff_applies_and_parses(self):
        out, manifest = self.pick(self.base, self.head, maximum=50)
        self.assertEqual(len(manifest["mutants"]), 7)
        for mutant in manifest["mutants"]:
            diff = os.path.join(out, mutant["hash"] + ".diff")
            run("git", "apply", "--check", diff, cwd=self.repo)
            run("git", "apply", diff, cwd=self.repo)
            run("gofmt", "-e", "-l", "pkg/clamp.go", cwd=self.repo)
            run("git", "apply", "-R", diff, cwd=self.repo)

    def test_the_same_range_picks_the_same_mutants(self):
        first = [mutant["hash"] for mutant in self.pick(self.base, self.head)[1]["mutants"]]
        second = [mutant["hash"] for mutant in self.pick(self.base, self.head)[1]["mutants"]]
        self.assertEqual(first, second)

    def test_a_range_that_only_touches_tests_yields_nothing(self):
        self.write("pkg/clamp_test.go", test_go.replace("5, 3", "7, 3"))
        self.commit("test only")
        head = run("git", "rev-parse", "HEAD", cwd=self.repo).strip()
        self.assertEqual(self.pick(self.head, head)[1]["mutants"], [])

    def test_a_pure_deletion_yields_nothing(self):
        self.write("pkg/clamp.go", head_go.replace("\tmessage := \"if a < b && c == 7\"\n\t_ = message\n", ""))
        self.commit("delete")
        head = run("git", "rev-parse", "HEAD", cwd=self.repo).strip()
        self.assertEqual(self.pick(self.head, head)[1]["mutants"], [])


if __name__ == "__main__":
    unittest.main()
