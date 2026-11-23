#!/usr/bin/env python3
"""The gates an Adamic head owes before `ahra loom submit`, and the evidence Loom's jobs hold for each: a checklist.

Usage: submit-gates.py --head <sha> --main <sha> --repository <path> --jobs <rows.json> --task <id>
                       [--live-hole] [--gates <directory>] [--json]

It is advisory (Adamic's ruling on #v3q9ayd, Nov 23 14:33Z). It never authorizes a submit: a gate it shows as met is one
the accepter still checks, and the accepter, Compiler or Adamic, is the authority, whose accept stays a comment on the
head's task. Its output says so in one line. A script measured inside the head's own tree can't be the authority, since
the head can always reach an input the judge doesn't list, so it reports what it found and plainly names what it
couldn't prove. Every lane head carries its printout in the accept comment.

It reads Compiler's map, cloud/gates/diff-to-gates.json, at main's tip (#mbtv3vv section 3). Every rule whose trigger
matches the head's diff against main fires, and the head owes the union of what the fired rules require: gates, each a
fixed recipe cloud/gates/<gate>.sh run as a Loom job, and same-diff rows, read from the diff itself. It shows evidence
for a gate only from a job row that is fresh and green by the map's definitions: its command's sha256 is the recipe's
read at main's tip, its commit is the head's own sha, it exited 0, and a paired gate's base, read from the verdict line
the recipe ends on, is main's tip or a main commit with no lowering-side change since. The rows are what
`ahra loom job <id> --json` prints, as a list, an object holding `jobs`, or one a line. It reads only git and that file:
nothing is run or asked.

It shows evidence only where it checks it by exact equality: a recipe's command sha256, the repository, the head's own
sha, an exit code, or a fixed line the recipe itself prints. A trigger fires on any byte change to the files it reads,
since a needless gate costs a job and a missed one a red block (the map's own words). What it can't read exactly is
owed by hand, printed as what the accepter checks.

It prints every owed gate and same-diff row with its evidence, or with why none was found (and for a gate the exact
`ahra loom job` line that asks it), then a verdict naming every one it couldn't prove. Its exit code reports and gates
nothing: 0 when it found evidence for every owed gate and row, 1 when it didn't (a checklist with items open, never a
refusal), and 2 when the head, the map or the rows can't be read.

A head is one change's diff against main's tip: a merge of main into it is still one head, and two heads merged together
are two, which the report says. Several --head arguments are each reported on their own.

mutants-red holds one job per mutant the head's *.mutants.md declares, each its own command, cloud/gates/mutants-red.sh
with that mutant written in, red at the head by the line the recipe prints only when the mutant's own test failed. An
adaptation's JavaScript is proved by identity or by value, the map's thenIf: stock-js-identical, or both
value-count-reached and adapter-output-test, and the report says which side, if either, it found.

The map, the recipes, steps.sh and units-compare.py are read at main's tip, never at the head, and the script stops
(exit 2) unless its own bytes are main's copy of it. A head that changes any judging code (JudgePaths) owes
gates-changed by hand, since its jobs source those files from its own tree. Where main holds no judging code yet,
--gates names the directory to read instead, and the report says this run's code is the caller's (cloud/gates/README.md).

--live-hole is the map's two-hour clock: a head naming a live-hole task submits on format, compilercensus,
parallelcensus and lower-whole, and every other gate it owes is printed as owed today. The caller vouches for the clock
and the pins, which the script can't see, and its report says so. Its same-diff rows are still listed, since a row can
only ride the head's own commit.
"""

import argparse
from datetime import datetime
import fnmatch
import hashlib
import json
from pathlib import Path
import re
import shlex
import subprocess
import sys


MapPath = "cloud/gates/diff-to-gates.json"
MapVersion = 5
ShaPattern = re.compile(r"[0-9a-f]{40}\Z")

# The map's definitions this script reads, word for word. A changed definition is a changed rule, so the script refuses
# a map whose words differ rather than read new words by old code; each is read by the function named beside it.
DefinitionsRead = {
    "productionGo": "a .go file not ending _test.go",
    "compilerProduction": "productionGo under internal/lower/, internal/ir/, internal/flow/, internal/native/ (outside "
                          "internal/native_runtime/), internal/javascript/, internal/load/, cmd/adamic/",
    "fresh": "a job counts only when its commit is the head's own sha, and its base (for a paired job) is main's tip at "
             "submit, or a main commit with no change under internal/lower, internal/ir, internal/flow, internal/native "
             "or internal/load since",
    "green": "the job's outcome is exit 0, and its command's sha256 equals the sha256 of the named gate recipe read at "
             "the head (Loom keeps a command's hash, never its text, so a gate is a fixed recipe file, "
             "cloud/gates/<gate>.sh, and the script compares hashes)",
    "red": "for a mutant job: exit non-zero, with the named test's FAIL line in its kept tail",
    "goFunctionBodyChanged": "the named function's body differs between base and head by go/ast, comments and "
                             "formatting ignored",
    # Read as the note on a set's report: the script judges one head, so a stacked set's combined census is the set's
    # own proof, never checked here.
    "freshStack": "for several adaptation heads submitted together (Seats' stacked form, Compiler on #bbssdcy Nov 23): "
                  "each head carries its own unit census at a recent main, 0 newly stopping; and the set carries one "
                  "combined census at main's tip at submit, base the tip, head every change applied, read by unit and by "
                  "site: 0 units newly stopping, every cleared unit attributable to a named head, the added sites equal "
                  "to the union of each head's own added sites, and any unit in two heads' cleared lists or holding an "
                  "extra site censused for those heads alone. If lower moves during the combined job it is re-asked "
                  "once, then the set submits on that green",
}
CompilerProductionPrefixes = ("internal/lower/", "internal/ir/", "internal/flow/", "internal/native/",
                              "internal/javascript/", "internal/load/", "cmd/adamic/")
# fresh's directories: a change under one of them between a paired job's base and main's tip makes the job stale.
FreshDirectories = ("internal/lower", "internal/ir", "internal/flow", "internal/native", "internal/load")

# Gates compared against a base: their recipe reads HEAD^ and ends on its verdict line naming it, which fresh holds to
# main's tip.
PairedGates = ("admission-delta", "unit-census-pair")
# Outcomes that are no verdict on the head: Loom's or the machine's failure, an end someone asked for, a job replaced,
# and a job not ended yet. Every other outcome is a verdict, and only Exit 0 is green.
NotVerdicts = ("Infrastructure", "Cancelled", "Superseded", None)
# Gates no fixed recipe proves, each with what the accepter checks. A job row never stands in for one: three review
# rounds found a label, a repeated job or a selector standing in for the proof. The script names what to check; the
# accepter checks it.
ByHandGates = {
    "value-count-reached": "the adaptation's value count job: 0 counterexamples, and calls observed above 0 on the "
                           "changed path, or a fixture held to Node that drives it",
    "stock-js-identical": "tsc's emit of the adapted files byte for byte against stock's; no tool on main compares them "
                          "yet",
}
# Same-diff rows owed by hand, wholly or for the part the script can't read, each with what the accepter checks.
ByHandRows = {
    "skip-census-row": "the skip-census row for each new skip, in adamic-loom's judge (census-extra.json), outside "
                       "adamic's tree",
    "pass-ledger-row": "that the pass ledger names each pass the change adds, a call added to the pass lists by go/ast, "
                       "which the script can't read",
    "statecopy-row": "that irSpecFields names each new exported internal/ir field, by go/ast, which the script can't read",
    "adapter-output-test": "that test.cjs runs the edited output, on a tree stage3/apply.sh wrote with its -before and "
                           "-after refs, and goes red on a reviewer's mutation of it (a swapped pair, a dropped entry)",
    "adapter-structural-pins": "that each digest or getText( pins a node the adapter edits, never a whole file, and that "
                               "its tests read its own -after snapshot, never the final applied tree or a constant "
                               "digest: whether it is a node is its receiver's type, which the script can't read",
    "gates-changed": "the change to the judging code the head's own jobs run (a recipe, steps.sh, units-compare.py, the "
                     "map, sandboxed-job.sh, this script, or the census tooling unit-census-pair.sh runs), read whole, "
                     "since a green of them is the head judging itself",
}
# The mutants-red recipe names its mutant on a line the script writes into the command it asks, so each mutant's
# command, and its sha256, is its own. The recipe holds the placeholder once.
MutantPlaceholder = "@MUTANT@"
# A mutant's name in its command: its declaration file and its row there, 1 first.
MutantName = re.compile(r"[A-Za-z0-9._/-]+\.mutants\.md:[1-9][0-9]*\Z")
# A machine-run mutant's cells: an exact patch in the head's tree, its one test, and the command that runs it.
MutantPatchCell = re.compile(r"[A-Za-z0-9._/-]+\.patch\Z")
MutantTestCell = re.compile(r"[A-Za-z_][A-Za-z0-9_/.-]*\Z")
MutantRunCell = re.compile(r"[^`\n]+\Z")
# The judging code: what the recipes source from the tree they run in, and this script. A head that changes any of it
# owes gates-changed by hand, since a job at the head runs the head's copy whatever recipe it was asked with.
JudgePaths = (
    "cloud/gates/",  # the map, every recipe, steps.sh and units-compare.py, which unit-census-pair.sh sources and runs
    "cloud/submit-gates.py",
    "cloud/sandboxed-job.sh",  # every recipe sources it; unit-census-pair.sh calls its sandboxed_npm and sandboxed_typescript_source
    # The census tooling: a head that changes what the census measures changes the judge's evidence (Seats, round 6)
    "stage3/census/",  # make_overlay.py and the go run tools it calls (statecopy, statementrewrite, refusalrewrite,
                       # importclean), and ./stage3/census/latent/tool, which unit-census-pair.sh builds and runs
    "stage3/apply.sh",  # unit-census-pair.sh runs it on each side
    "stage3/apply.py",  # apply.sh execs it
    "stage3/caches.py",  # apply.py imports it
    "stage3/meter/entry_overlay.py",  # unit-census-pair.sh runs it for the entry pass's overlay
    "stage3/api/package.json",  # sandboxed_npm stage3/api runs npm ci there, which installs from these two
    "stage3/api/package-lock.json",
    # stage3/apply.py also runs stage3/adapt/00-setup/adapt.cjs and every adapter, but those are what an adaptation
    # head changes and the census measures, never the judge's own code
)
ScriptPath = "cloud/submit-gates.py"
# The live-hole exception's gates (the map's exceptions, two-hour-clock); its pins ride the line, never a recipe.
LiveHoleGates = ("format", "compilercensus", "parallelcensus", "lower-whole")

PassLedgerPath = "stage3/census/latent/pass_ledger.go.txt"
LedgerPath = "stage3/upstream/LEDGER.md"
# The map names statecopy/main.go, but irSpecFields lives beside it (ir_fields.go on main 96a1291b), so the whole
# directory is read.
StatecopyDirectory = "stage3/census/latent/statecopy/"
PendingMessage = re.compile(r"awaits #[0-9a-z]{6,7}: [^\n]*\S")
# A program's expected outcome, beside it by its stem (x.stop) or by its whole name (x.a.stop, the dead-zone pins'
# convention, internal/oracle/dead_zone_stops_test.go); `lowers` there means it lowers and runs, never a stop.
ExpectedStopSidecars = (".stop", ".stopped")
LowersOutcome = "lowers"
# The line a recipe ends on (cloud/gates/steps.sh, gate_end), read only as the tail's last line and whole.
VerdictLine = re.compile(r"VERDICT (PASS|FAIL) (\S+) head ([0-9a-f]{40}|unknown) base ([0-9a-f]{40}|none)\Z")


class Unreadable(Exception):
    """The head, the map or the rows can't be read; exit 2."""


class Repository:
    def __init__(self, root):
        self.root = Path(root)
        self.blobs = {}

    def git(self, *arguments, check=True):
        result = subprocess.run(["git", "-C", str(self.root), *arguments], capture_output=True)
        if check and result.returncode != 0:
            raise Unreadable(f"git {' '.join(arguments)}: {result.stderr.decode(errors='replace').strip()}")
        return result

    def text(self, *arguments):
        return self.git(*arguments).stdout.decode(errors="replace")

    def commit(self, revision):
        found = self.git("rev-parse", "--verify", "-q", revision + "^{commit}", check=False)
        if found.returncode != 0:
            raise Unreadable(f"{revision} is no commit in {self.root}")
        return found.stdout.decode().strip()

    def show(self, revision, path):
        """A file's bytes at a revision, or None where it doesn't exist."""
        key = (revision, path)
        if key not in self.blobs:
            found = self.git("cat-file", "-p", f"{revision}:{path}", check=False)
            self.blobs[key] = found.stdout if found.returncode == 0 else None
        return self.blobs[key]

    def files(self, revision, directory):
        rows = self.git("ls-tree", "-z", "--name-only", revision, "--", directory.rstrip("/") + "/").stdout
        return [name.decode() for name in rows.split(b"\0") if name]

    def ancestor(self, older, newer):
        return self.git("merge-base", "--is-ancestor", older, newer, check=False).returncode == 0


# Paths.

def production_go(path):
    return path.endswith(".go") and not path.endswith("_test.go")


def compiler_production(path):
    return production_go(path) and path.startswith(CompilerProductionPrefixes)


def glob_pattern(pattern):
    """A map glob as a regular expression: ** spans directories, * and ? stay in one, {a,b} is either."""
    expression, index = "", 0
    while index < len(pattern):
        if pattern.startswith("**/", index):
            expression, index = expression + "(?:.*/)?", index + 3
        elif pattern.startswith("**", index):
            expression, index = expression + ".*", index + 2
        elif pattern[index] == "*":
            expression, index = expression + "[^/]*", index + 1
        elif pattern[index] == "?":
            expression, index = expression + "[^/]", index + 1
        elif pattern[index] == "{":
            end = pattern.index("}", index)
            expression += "(?:" + "|".join(re.escape(choice) for choice in pattern[index + 1:end].split(",")) + ")"
            index = end + 1
        else:
            expression, index = expression + re.escape(pattern[index]), index + 1
    return re.compile(expression + r"\Z")


def path_matcher(spec):
    """The paths a map spec names: a definition, `productionGo under <prefix>`, or a glob."""
    if spec == "compilerProduction":
        return compiler_production
    if spec == "productionGo":
        return production_go
    under = re.fullmatch(r"productionGo under (\S+/)", spec)
    if under:
        prefix = under.group(1)
        return lambda path: production_go(path) and path.startswith(prefix)
    if re.search(r"\s", spec):
        raise Unreadable(f"the map's path spec {spec!r} is neither a definition, productionGo under a directory, nor a glob")
    expression = glob_pattern(spec)
    return lambda path: expression.match(path) is not None


# Go source. This script is standard library Python, and a head's diff is read on a Mac that builds no Go, so it has no
# go/ast. A pass list is never located in its file: three rounds of review found a token compare, then a text compare,
# then a declaration's line steered by a comment, each selecting the wrong span, so any change to a file holding a pass
# list fires (trigger_function_bodies). Struct fields are read by tokens (struct_fields), with Go's identifiers Unicode
# as the spec has them, and fire when unsure.

GoToken = re.compile(r"""
    (?P<space>\s+)
  | (?P<line>//[^\n]*)
  | (?P<block>/\*.*?\*/)
  | (?P<raw>`[^`]*`)
  | (?P<string>"(?:[^"\\\n]|\\.)*")
  | (?P<rune>'(?:[^'\\\n]|\\.)+')
  | (?P<word>[^\W\d]\w*)
  | (?P<number>[0-9][0-9A-Za-z_.]*)
  | (?P<operator>:=|\.\.\.|<-|&&|\|\||==|!=|<=|>=|<<|>>|&\^|\+\+|--|[-+*/%&|^]=?|<<=|>>=|.)
""", re.VERBOSE | re.DOTALL)


# Go's keywords, and the four of them after which a line end is a semicolon (the spec's semicolon rule 1).
GoKeywords = {"break", "case", "chan", "const", "continue", "default", "defer", "else", "fallthrough", "for", "func",
              "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct",
              "switch", "type", "var"}
GoLineEndKeywords = {"break", "continue", "fallthrough", "return"}
GoLineEndOperators = {"++", "--", ")", "]", "}"}


def ends_statement(token):
    """Whether a line end after this token is a semicolon, by Go's rule: after an identifier, a literal, one of break,
    continue, fallthrough and return, or ++, --, ), ] or }."""
    kind, text = token[0], token[1]
    if kind == "word":
        return text not in GoKeywords or text in GoLineEndKeywords
    return kind in ("number", "string", "rune", "raw") or text in GoLineEndOperators


def go_tokens(source):
    """Each token as (kind, text, line), comments and whitespace dropped and Go's automatic semicolons inserted, each
    as the operator `;` an explicit one is, on its statement's line. Without them `return f()` and a bare `return` then
    `f()` read as the same tokens, and they're one statement against two (review-r's must 1)."""
    tokens, line = [], 1
    for match in GoToken.finditer(source):
        kind, text = match.lastgroup, match.group()
        if kind in ("space", "line", "block"):
            # A newline, or a general comment holding one, ends the line
            if "\n" in text and tokens and tokens[-1][1] != ";" and ends_statement(tokens[-1]):
                tokens.append(("operator", ";", tokens[-1][2]))
        else:
            tokens.append((kind, text, line))
        line += text.count("\n")
    if tokens and tokens[-1][1] != ";" and ends_statement(tokens[-1]):
        tokens.append(("operator", ";", tokens[-1][2]))
    return tokens


def matching_brace(tokens, opening):
    """The index of the brace that closes tokens[opening], or None."""
    depth = 0
    for index in range(opening, len(tokens)):
        if tokens[index][1] == "{":
            depth += 1
        elif tokens[index][1] == "}":
            depth -= 1
            if depth == 0:
                return index
    return None


# A word that a call could name, followed by its opening parenthesis: what the pass ledger row lists for its owner's read,
# never what holds the row.
GoCallee = re.compile(r"([^\W\d]\w*)\s*\(")


def struct_fields(source, wanted):
    """Each struct type the source declares at the top level, alone (`type T struct {`) or in a `type (...)` group, as
    name to its field names (an embedded field's name is its type's), for the names wanted(name) accepts. None when a
    struct's braces don't close: unsure."""
    tokens = go_tokens(source)
    candidates, braces, group = [], 0, None
    for index, token in enumerate(tokens):
        first = index == 0 or tokens[index - 1][2] != token[2]
        if token[1] == "{":
            braces += 1
        elif token[1] == "}":
            braces -= 1
        elif braces == 0 and token[1] == "type" and first and index + 1 < len(tokens):
            if tokens[index + 1][1] == "(":
                group = 0
            else:
                candidates.append(index + 1)
        elif group is not None and braces == 0:
            if token[1] == "(":
                group += 1
            elif token[1] == ")":
                group -= 1
                if group == 0:
                    group = None
            elif group == 1 and first and token[0] == "word":
                candidates.append(index)
    structs = {}
    for index in candidates:
        name, cursor = tokens[index][1], index + 1
        if cursor < len(tokens) and tokens[cursor][1] == "[" and tokens[cursor][2] == tokens[index][2]:  # type parameters
            depth = 0
            while cursor < len(tokens):
                depth += {"[": 1, "]": -1}.get(tokens[cursor][1], 0)
                cursor += 1
                if depth == 0:
                    break
        if cursor + 1 >= len(tokens) or tokens[cursor][1] != "struct" or tokens[cursor + 1][1] != "{" or not wanted(name):
            continue
        closing = matching_brace(tokens, cursor + 1)
        if closing is None:
            return None
        fields, line, depth = [], [], 0
        for token in tokens[cursor + 2:closing] + [("operator", ";", -1)]:
            if line and token[2] != line[-1][2] and depth == 0:
                fields.extend(field_names(line))
                line = []
            if token[1] in ("{", "(", "["):
                depth += 1
            elif token[1] in ("}", ")", "]"):
                depth -= 1
            if token[2] != -1:
                line.append(token)
        structs[name] = fields
    return structs


def field_names(line):
    """The names one field line declares: `A, B T` gives A and B, an embedded `*pkg.T` or `pkg.T` gives T."""
    tokens = [token for token in line if token[0] not in ("raw", "string")]  # a tag is a field's last literal
    names, cursor = [], 0
    while cursor < len(tokens) and tokens[cursor][0] == "word":
        names.append(tokens[cursor][1])
        cursor += 1
        if cursor < len(tokens) and tokens[cursor][1] == ",":
            cursor += 1
            continue
        break
    if names and cursor < len(tokens) and tokens[cursor][1] != ".":
        return names
    words = [token[1] for token in tokens if token[0] == "word"]
    return words[-1:]


# The head's diff.

class Diff:
    """The head's own change: the diff from its merge base with main to the head."""

    def __init__(self, repository, head, main):
        self.repository, self.head, self.main = repository, head, main
        self.base = repository.text("merge-base", main, head).strip()
        self.commits = repository.text("rev-list", "--first-parent", f"{self.base}..{head}").split()
        statuses = repository.git("diff", "--no-renames", "--name-status", "-z", self.base, head).stdout.decode().split("\0")
        self.status = {}
        for index in range(0, len(statuses) - 1, 2):
            self.status[statuses[index + 1]] = statuses[index]
        self.added_lines, self.added_numbered = {}, {}
        current, number = None, 0
        patch = repository.git("diff", "--no-renames", "-U0", "--no-color", "--src-prefix=a/",
                             "--dst-prefix=b/", self.base, head).stdout.decode(errors="replace")
        for line in patch.splitlines():
            if line.startswith("+++ "):
                current = line[6:] if line.startswith("+++ b/") else None
            elif line.startswith("@@ "):
                number = int(re.match(r"@@ -\S+ \+(\d+)", line).group(1))
            elif line.startswith("+") and current is not None:
                self.added_lines.setdefault(current, []).append(line[1:])
                self.added_numbered.setdefault(current, []).append((number, line[1:]))
                number += 1

    def paths(self):
        return sorted(self.status)

    def added(self):
        return sorted(path for path, status in self.status.items() if status == "A")

    def source(self, revision, path):
        data = self.repository.show(revision, path)
        return None if data is None else data.decode(errors="replace")


# Triggers. Each returns (fired, why) where why names what matched, or None when unsure and it fired anyway.

def trigger_paths(diff, trigger):
    matcher = path_matcher(trigger["paths"])
    paths = [path for path in diff.paths() if matcher(path)]
    if "addedLineMatches" not in trigger:
        return (bool(paths), f"{len(paths)} paths under {trigger['paths']}, first {paths[0]}" if paths else "")
    expression = re.compile(trigger["addedLineMatches"])
    for path in paths:
        for line in diff.added_lines.get(path, []):
            if expression.search(line):
                return True, f"{path} adds {line.strip()[:100]!r}"
    return False, ""


def pass_list_files(diff, trigger):
    """Each changed file that may hold a named pass list, with why: the file the map names, on any change at all, and any
    production Go file of its package whose bytes at the base or the head hold a named function's name as a word, since
    a function called by its bare name lives in its package's directory, wherever a head moves it. No span is selected,
    so nothing in a file, a comment or a string included, can hide a change from the trigger."""
    files = {}
    for named in trigger["goFunctionBodyChanged"]:
        path, name = (part.strip() for part in named.split(":", 1))
        if path in diff.status:
            files.setdefault(path, f"{path} changed, and the map names {name} there")
        word = re.compile(r"(?<!\w)" + re.escape(name) + r"(?!\w)")
        directory = path.rsplit("/", 1)[0]
        for other in diff.paths():
            if other == path or other in files or not production_go(other) or other.rsplit("/", 1)[0] != directory:
                continue
            if any(word.search(diff.source(revision, other) or "") for revision in (diff.base, diff.head)):
                files[other] = f"{other} changed and holds the word {name}"
    return files


def trigger_function_bodies(diff, trigger):
    """Fires on any change to a file that holds a named pass list (pass_list_files), byte for byte. The map's words are
    go/ast with formatting ignored; this reads more broadly, since a needless gate costs a job and a missed one a red
    block (the map's about)."""
    files = pass_list_files(diff, trigger)
    return bool(files), "; ".join(f"{why}: fires" for why in files.values())


def struct_scope(spec):
    """A goStructFieldAdded spec as (directory, struct name or None for every exported struct, exported fields only)."""
    named = re.fullmatch(r"(\S+): type (\w+) struct", spec)
    if named:
        return named.group(1), named.group(2), False
    if re.fullmatch(r"\S+: any exported struct, exported field", spec):
        return spec.split(":", 1)[0], None, True
    raise Unreadable(f"the map's goStructFieldAdded spec {spec!r} isn't one this script reads")


def new_struct_fields(diff, spec):
    """The fields added under the spec, as Struct.Field, or None when unsure."""
    directory, struct, exported = struct_scope(spec)
    touched = [path for path in diff.paths()
               if production_go(path) and path.rsplit("/", 1)[0] == directory]
    if not touched:
        return []
    wanted = (lambda name: name == struct) if struct else (lambda name: name[:1].isupper())
    sides = []
    for revision in (diff.base, diff.head):
        fields = {}
        for path in diff.repository.files(revision, directory):
            if not production_go(path) or path.rsplit("/", 1)[0] != directory:
                continue
            structs = struct_fields(diff.source(revision, path) or "", wanted)
            if structs is None:
                return None
            for name, names in structs.items():
                fields.setdefault(name, set()).update(names)
        sides.append(fields)
    added = []
    for name, names in sorted(sides[1].items()):
        for field in sorted(names - sides[0].get(name, set())):
            if not exported or field[:1].isupper():
                added.append(f"{name}.{field}")
    return added


def trigger_struct_field(diff, trigger):
    """Fires on any change to a production Go file in the directory the spec names, the files its structs are read
    from: whether a change adds a field is go/ast's question, and a token reader steered by a comment or a string would
    say no when the answer is yes, so the trigger never reads the fields at all."""
    directory = struct_scope(trigger["goStructFieldAdded"])[0]
    touched = [path for path in diff.paths() if production_go(path) and path.rsplit("/", 1)[0] == directory]
    return bool(touched), f"{len(touched)} production files in {directory} changed, first {touched[0]}: fires" if touched else ""


def expected_stop_programs(diff, pattern):
    matcher = path_matcher(pattern)
    return [path for path in diff.added() if matcher(path)]


def trigger_added_paths(diff, trigger):
    programs = expected_stop_programs(diff, trigger["addedPaths"])
    return bool(programs), f"{len(programs)} added: {', '.join(programs[:5])}" if programs else ""


def trigger_added_or_changed(diff, trigger):
    matcher = path_matcher(trigger["addedOrChangedPaths"])
    paths = [path for path, status in diff.status.items() if status != "D" and matcher(path)]
    return bool(paths), ", ".join(sorted(paths)[:5])


TriggerReaders = {
    "paths": trigger_paths,
    "goFunctionBodyChanged": trigger_function_bodies,
    "goStructFieldAdded": trigger_struct_field,
    "addedPaths": trigger_added_paths,
    "addedOrChangedPaths": trigger_added_or_changed,
}
# Keys a trigger may carry beside the one that names its reader.
TriggerModifiers = {"addedLineMatches", "withExpectedStop"}


def read_map(text):
    try:
        gates_map = json.loads(text)
    except ValueError as error:
        raise Unreadable(f"{MapPath} doesn't parse: {error}") from error
    if gates_map.get("version") != MapVersion:
        raise Unreadable(f"{MapPath} is version {gates_map.get('version')}; this script reads version {MapVersion}")
    if gates_map.get("definitions") != DefinitionsRead:
        changed = sorted(set(gates_map.get("definitions", {})) ^ set(DefinitionsRead) | {
            name for name in DefinitionsRead if gates_map.get("definitions", {}).get(name) != DefinitionsRead[name]})
        raise Unreadable(f"{MapPath}'s definitions differ from the ones this script reads ({', '.join(changed)}): "
                         "update the script with the map")
    for rule in gates_map["rules"]:
        for trigger in rule["triggers"]:
            readers = set(trigger) - TriggerModifiers
            if not (readers <= set(TriggerReaders) | {"always", "rule"} and len(readers) == 1):
                raise Unreadable(f"rule {rule['id']}'s trigger {trigger} isn't one this script reads")
            if "paths" in trigger:
                path_matcher(trigger["paths"])
            if "addedPaths" in trigger:
                path_matcher(trigger["addedPaths"])
            if "goStructFieldAdded" in trigger:
                struct_scope(trigger["goStructFieldAdded"])
        for gate in rule.get("requires", []) + rule.get("thenIf", {}).get("requires", []):
            if gate not in gates_map["gates"]:
                raise Unreadable(f"rule {rule['id']} requires {gate}, which the map's gates don't hold")
        for row in rule.get("sameDiff", []) + rule.get("thenIf", {}).get("sameDiff", []):
            if row not in gates_map["sameDiff"]:
                raise Unreadable(f"rule {rule['id']} requires the same-diff row {row}, which the map doesn't hold")
        if "thenIf" in rule:
            # The choice is read as one gate of the rule's own set false beside the branch it owes (choice_of)
            conditions = {key: value for key, value in rule["thenIf"].items() if isinstance(value, bool)}
            if list(conditions.values()) != [False] or next(iter(conditions)) not in rule.get("requires", []):
                raise Unreadable(f"rule {rule['id']}'s thenIf isn't one of its own required gates set false, the only "
                                 "choice this script reads")
    return gates_map


def fire(gates_map, diff):
    """Each fired rule's id and why, in the map's order, a rule trigger read after the rest."""
    fired = {}
    for rule in gates_map["rules"]:
        for trigger in rule["triggers"]:
            if "always" in trigger:
                fired.setdefault(rule["id"], "every head")
            elif "rule" not in trigger:
                reader = TriggerReaders[(set(trigger) - TriggerModifiers).pop()]
                matched, why = reader(diff, trigger)
                if matched:
                    fired.setdefault(rule["id"], why)
    for rule in gates_map["rules"]:
        for trigger in rule["triggers"]:
            if "rule" in trigger and trigger["rule"] in fired:
                fired.setdefault(rule["id"], f"rule {trigger['rule']} fired")
    return {rule["id"]: fired[rule["id"]] for rule in gates_map["rules"] if rule["id"] in fired}


# Same-diff rows.

def row_statecopy(diff, rule_ids):
    """Owed whenever internal/ir changes. Which exported fields are new is go/ast's question, so the row is owed by hand;
    the fields a token read finds new, and whether irSpecFields names each, are listed for the accepter's read."""
    added = new_struct_fields(diff, "internal/ir: any exported struct, exported field")
    text = "".join(diff.source(diff.head, path) or "" for path in diff.repository.files(diff.head, StatecopyDirectory)
                   if path.endswith(".go"))
    listed = ("a struct wasn't read whole" if added is None else
              ", ".join(field + (" (named)" if f'"{field}"' in text else " (not named)") for field in added) or "none")
    read = f"; the new fields a token read finds, for the read: {listed}"
    return False, owed_by_hand("statecopy-row") + read


def added_callees(diff):
    """Each word an added line of a pass list's file puts before a parenthesis, for the owner's read of the ledger row."""
    files = {}
    for rule in diff.gates_map["rules"]:
        if rule["id"] == "pass-added":
            for trigger in rule["triggers"]:
                if "goFunctionBodyChanged" in trigger:
                    files.update(pass_list_files(diff, trigger))
    callees = []
    for path in sorted(files):
        for line in diff.added_lines.get(path, []):
            callees += [word for word in GoCallee.findall(line) if word not in GoKeywords and word not in callees]
    return callees


def row_pass_ledger(diff, rule_ids):
    """Owed whenever the pass lists' files change. Whether a call was added is the map's go/ast compare, which this
    script can't make (a comment opened or closed by an added line turns unchanged lines into calls), so the row is owed
    by hand; the words that could name a call, and whether the ledger names each, are listed for the accepter's read."""
    ledger = diff.source(diff.head, PassLedgerPath) or ""
    callees = added_callees(diff)
    listed = [word + (" (named)" if 'Name: "' + word + '"' in ledger else " (not named)") for word in callees]
    read = f"; the added lines' callees, for the read: {', '.join(listed) or 'none'}"
    return False, owed_by_hand("pass-ledger-row") + read


def expected_outcome(diff, program):
    """The program's expected outcome from its sidecars, (path, text), or None when it has none. Every convention is
    read, and two sidecars that say different things are refused rather than one chosen: (paths, None)."""
    stem = program.rsplit(".", 1)[0]
    found = []
    for path in [stem + sidecar for sidecar in ExpectedStopSidecars] + [program + ".stop"]:
        text = diff.source(diff.head, path)
        if text is not None:
            found.append((path, text.strip()))
    if not found:
        return None
    if len({text for path, text in found}) > 1:
        return ", ".join(path for path, text in found), None
    return found[0]


def row_pending(diff, rule_ids):
    """Every added program that stops where Node runs carries a .pending naming a task, changed in this diff. A program's
    outcome is read from files alone, by exact name and text: its sidecar reading `lowers` lowers and runs, any other
    text is a stop, and a .refused beside it is a refusal. One with none of them is owed its sidecar: the commit
    message is never read for an outcome, since a word in it is no file the tree holds."""
    pattern = next(trigger["addedPaths"] for rule in diff.gates_map["rules"] if rule["id"] == "stopping-pin"
                   for trigger in rule["triggers"] if "addedPaths" in trigger)
    missing = []
    for program in expected_stop_programs(diff, pattern):
        stem = program.rsplit(".", 1)[0]
        outcome = expected_outcome(diff, program)
        if outcome is not None and outcome[1] is None:
            missing.append(f"{outcome[0]} disagree on {program}'s expected outcome")
            continue
        if outcome is not None:
            if outcome[1] == LowersOutcome:
                continue
            why = f"{outcome[0]} expects a stop"
        elif diff.repository.show(diff.head, stem + ".refused") is not None:
            continue  # a refusal, never a stop where Node runs
        else:
            missing.append(f"{program} has no expected-outcome sidecar: add {stem}.stop reading `{LowersOutcome}`, or "
                           "the stop it expects with its .pending")
            continue
        pending = diff.source(diff.head, stem + ".pending")
        if stem + ".pending" not in diff.status or diff.status[stem + ".pending"] == "D":
            missing.append(f"{why}, and {stem}.pending isn't added or changed in this diff")
        elif pending is None or not PendingMessage.fullmatch(pending.strip()):
            missing.append(f"{why}, and {stem}.pending doesn't read `awaits #<task>: <reason>`")
    return not missing, "; ".join(missing) or "every added program's outcome is read"


def row_ledger(diff, rule_ids):
    if LedgerPath in diff.status:
        return True, f"{LedgerPath} changed"
    return False, f"{LedgerPath} isn't in the head's diff"


def row_mutant_evidence(diff, rule_ids):
    """Held when the head's own *.mutants.md files declare at least one mutant and every one is in the machine form
    (declared_mutants), each naming in its Red cell the test that goes red with the change undone. That a name is there
    is read exactly; that the test goes red is mutants-red's, which the same file fires."""
    declared = declared_mutants(diff)
    if not declared:
        return False, "the head's paths hold no *.mutants.md naming the test that goes red with the change undone"
    unread = [mutant["why"] for mutant in declared if mutant["test"] is None]
    if unread:
        return False, "; ".join(unread)
    return True, "; ".join(f"{mutant['name']} names {mutant['test']}" for mutant in declared)


AdaptDirectory = re.compile(r"(stage3/adapt/[^/]+)/")
AdapterCode = re.compile(r"stage3/adapt/.+\.(?:cjs|js|mjs)\Z")
DigestUse = re.compile(r"createHash|sha256|digest", re.IGNORECASE)
# A module of code an adapter requires from its own tree, where a digest or a pin can live out of the adapter's sight.
LocalCodeRequire = re.compile(r"""require\(\s*['"]\.{1,2}/[^'"]*?(?<!\.json)['"]\s*\)""")
# A digest's input that is a whole file's text, by how adapters read one: the file read from disk, a source file's
# whole text, or its full text. Read only to refuse, never to hold.
WholeFileText = re.compile(r"readFileSync|fs\.read|getFullText\s*\(|\bsourceFile\.text\b|\bfile\.text\b|\bsource\.text\b"
                           r"|getSourceFile\s*\([^()]*\)\s*\.\s*(?:getText\s*\(|text\b)")
# A name the adapter binds to a TypeScript SourceFile, which is a whole file: its getText( or text is the file's text.
SourceFileBinding = re.compile(r"(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=[^;\n]*\b(?:createSourceFile|getSourceFile)\s*\(")


def digest_inputs(text):
    """Each digest input an adapter's source holds: the argument of every `.update(` and of every call to a function it
    defines that makes a digest, the helpers' own definitions taken out first, since a helper's parameter says nothing
    of what its callers hash. An argument is read to its closing parenthesis on its own line."""
    helpers = set()
    for found in re.finditer(r"^.*(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=[^;\n]*createHash.*$", text, re.MULTILINE):
        helpers.add(found.group(1))
    for found in re.finditer(r"function\s+([A-Za-z_$][\w$]*)\s*\([^)]*\)\s*\{[^}]*createHash[^}]*\}", text):
        helpers.add(found.group(1))
    body = re.sub(r"^.*(?:const|let|var)\s+[A-Za-z_$][\w$]*\s*=[^;\n]*createHash.*$", "", text, flags=re.MULTILINE)
    body = re.sub(r"function\s+[A-Za-z_$][\w$]*\s*\([^)]*\)\s*\{[^}]*createHash[^}]*\}", "", body)
    calls = r"\.update\s*\(" + (r"|(?<![\w$.])(?:" + "|".join(map(re.escape, sorted(helpers))) + r")\s*\(" if helpers else "")
    inputs = []
    for match in re.finditer(calls, body):
        depth, cursor = 1, match.end()
        while cursor < len(body) and depth and body[cursor] != "\n":
            depth += {"(": 1, ")": -1}.get(body[cursor], 0)
            cursor += 1
        inputs.append(body[match.end():cursor - 1].strip())
    return inputs


def whole_file_input(argument, source_files):
    """Whether a digest's input is, as far as the script can see, a whole file's text: one read whole, or a SourceFile's
    text through a name the adapter bound to one. True refuses; False holds nothing, since an input the script can't see
    may still be a whole file."""
    if WholeFileText.search(argument):
        return True
    return any(re.search(r"(?<![\w$.])" + re.escape(name) + r"\s*\.\s*(?:getText\s*\(|text\b)", argument)
               for name in source_files)


def row_structural_pins(diff, rule_ids):
    """Each changed adapter's code pins its input by the nodes it edits, never by a whole file's digest. A pin is never
    held by the machine: whether getText( reads a node or a whole file is its receiver's type, which no pattern reads. A
    digest the script can see is of a whole file is reported as one; any other digest, any getText(, or a module of code
    the adapter requires from its tree is owed by hand. Code holding none of those shows no pin, which the accepter still
    checks, since a pin can live where no pattern reads it (review-r round 4). test.cjs is the output test's."""
    refused, pinned = [], []
    for path, status in sorted(diff.status.items()):
        if status == "D" or not AdapterCode.match(path) or path.rsplit("/", 1)[-1] == "test.cjs":
            continue
        text = diff.source(diff.head, path) or ""
        source_files = set(SourceFileBinding.findall(text))
        refused += [f"{path} hashes {argument!r}, a whole file" for argument in digest_inputs(text)
                    if whole_file_input(argument, source_files)]
        holds = [what for what, found in (("a digest", DigestUse.search(text)), ("getText(", "getText(" in text),
                                          ("a module of its own code", LocalCodeRequire.search(text))) if found]
        if holds:
            pinned.append(f"{path} holds {', '.join(holds)}")
    if refused:
        return False, "; ".join(refused)
    if not pinned:
        return True, "no changed adapter code makes a digest, reads getText( or requires its own code"
    return False, "; ".join(pinned) + "; " + owed_by_hand("adapter-structural-pins")


def row_output_test(diff, rule_ids):
    """Each changed adapter's directory carries test.cjs, and the accepter checks it goes red on a mutation of the output."""
    directories = sorted({found.group(1) for path, status in diff.status.items() if status != "D"
                          for found in [AdaptDirectory.match(path)] if found})
    absent = [directory for directory in directories if diff.repository.show(diff.head, directory + "/test.cjs") is None]
    if absent:
        return False, f"no test.cjs beside the adapter in {', '.join(absent)}"
    return False, f"test.cjs is in {', '.join(directories)}; " + owed_by_hand("adapter-output-test")


def row_skip_census(diff, rule_ids):
    skips = "; the new skips: " + ", ".join(added_skips(diff))
    return False, owed_by_hand("skip-census-row") + skips


RowReaders = {
    "skip-census-row": row_skip_census,
    "adapter-structural-pins": row_structural_pins,
    "adapter-output-test": row_output_test,
    "statecopy-row": row_statecopy,
    "pass-ledger-row": row_pass_ledger,
    "pending-sidecar": row_pending,
    "ledger-line": row_ledger,
    "mutant-evidence-path": row_mutant_evidence,
}


# Jobs.

def read_rows(path, kind="job", key="jobs"):
    """Rows from a file, as a list, an object holding them under key, or one a line, each with an id: Loom's job rows,
    or a task's comment rows."""
    with open(path, newline="") as file:
        text = file.read()
    try:
        data = json.loads(text)
    except ValueError:
        try:
            data = [json.loads(line) for line in text.splitlines() if line.strip()]
        except ValueError as error:
            raise Unreadable(f"{path} holds neither JSON nor one JSON row a line: {error}") from error
    if isinstance(data, dict):
        data = data.get(key, [data] if "id" in data else None)
    if not isinstance(data, list) or not all(isinstance(row, dict) and isinstance(row.get("id"), str) for row in data):
        raise Unreadable(f"{path} holds no list of {kind} rows, each with an id")
    return data


def recipe_command(text):
    """The command a recipe's bytes ask: its text with its trailing newlines taken off, as `"$(git show ...)"` passes it,
    since a shell's command substitution strips them."""
    return text.rstrip("\n")


def recipe_hash(text):
    return hashlib.sha256(recipe_command(text).encode()).hexdigest()


def ask_flags(text):
    """The flags a recipe's `# ask:` line adds to its ask (a machine size, a ceiling)."""
    found = re.search(r"^# ask: (.*)$", text, re.MULTILINE)
    flags = shlex.split(found.group(1)) if found else []
    for index in range(0, len(flags), 2):
        if flags[index] not in ("--size", "--ceiling", "--about") or index + 1 >= len(flags):
            raise Unreadable(f"a recipe's ask line {found.group(0)!r} holds {flags[index]!r}; it takes --size, --ceiling and --about")
    return flags


def fresh_base(repository, base, main):
    """Why a paired job's base isn't fresh, or None when it is."""
    if base == main:
        return None
    if not repository.ancestor(base, main):
        return f"its base {base[:8]} isn't a main commit"
    moved = repository.text("diff", "--name-only", base, main, "--", *FreshDirectories).split()
    if moved:
        return f"its base {base[:8]} is {len(moved)} lowering-side files behind main's tip ({moved[0]}, ...)"
    return None


def ended_at(row):
    """A row's end as an instant, or None when it has none that reads. Loom writes 2026-11-23T09:17:31.241Z; parsed, never
    compared as text, so a precision Loom someday drops can't reorder two ends. One with no zone names no instant, so it
    reads as none: compared with a zoned one it would raise, and guessed it could reorder."""
    try:
        instant = datetime.fromisoformat(str(row.get("endedAt")).replace("Z", "+00:00"))
    except ValueError:
        return None
    return instant if instant.tzinfo is not None else None


def final_verdict(tail):
    """The tail's last line as a recipe's verdict, (outcome, gate, head, base), or None when it isn't one whole: an
    earlier verdict line followed by other output is no verdict, since the job went on after it."""
    lines = (tail or "").rstrip().splitlines()
    found = VerdictLine.match(lines[-1].strip()) if lines else None
    return found.groups() if found else None


def newest_runs(rows, wanted, head):
    """(the newest runs of a command at the head that reached a verdict, why) or (None, why). A run that ended with no
    verdict of its own (NotVerdicts) is passed over; any other outcome is a verdict. Runs that end at the same instant
    are all returned, in job id order, for the caller to hold every one."""
    ran = [row for row in rows if row.get("atRepository") == "adamic" and row.get("commandSha256") == wanted]
    why = []
    verdicts = []
    for row in sorted(ran, key=lambda row: row["id"]):
        if row.get("atSha") != head:
            why.append(f"{row['id']} ran the command at {str(row.get('atSha'))[:8]}, not the head")
        elif row.get("outcome") in NotVerdicts:
            why.append(f"{row['id']} at the head ended {row.get('outcome') or 'not yet'}, no verdict")
        elif ended_at(row) is None:
            return None, f"{row['id']} at the head has no end that reads ({row.get('endedAt')!r}), so no run is newest"
        else:
            verdicts.append(row)
    if not verdicts:
        return None, "; ".join(why + ([] if ran else ["no job ran the command"]))
    last = max(ended_at(row) for row in verdicts)
    return [row for row in verdicts if ended_at(row) == last], ""


def judge_gate(gate, recipe, rows, head, main, repository):
    """(job id or None, why) for one recipe gate.

    The newest run of the recipe at the head that reached a verdict decides, never the newest green: a green and then a
    red at the same sha is a red, since a flake is code until shown otherwise. Only exit 0 is green, and runs that end at
    the same instant are all read, any one not green refusing the gate. A paired gate's verdict is the tail's final line,
    whole: its head must be the job's, its base the head's parent (the base the recipe measured, HEAD^), and that base
    fresh.
    """
    newest, why = newest_runs(rows, recipe_hash(recipe), head)
    if newest is None:
        titled = [row["id"] for row in rows if row.get("atSha") == head and gate in (row.get("title") or "")
                  and row.get("commandSha256") != recipe_hash(recipe)]
        return None, why + (f"; {', '.join(titled)} at the head name it in a title, but their command isn't the recipe"
                            if titled else "")
    parent = repository.text("rev-parse", "--verify", "-q", head + "^").strip() if gate in PairedGates else None
    for row in newest:
        if row.get("outcome") != "Exit" or row.get("exit") != 0:
            return None, f"{row['id']}, the newest verdict at the head, ended {row.get('outcome')} {row.get('exit')}"
        if gate in PairedGates:
            verdict = final_verdict(row.get("tailText"))
            if verdict is None or verdict[0] != "PASS" or verdict[1] != gate:
                return None, f"{row['id']}, the newest verdict at the head, doesn't end on a passing verdict line for {gate}"
            if verdict[2] != row.get("atSha"):
                return None, f"{row['id']}'s verdict line names the head {verdict[2][:8]}, not the job's {head[:8]}"
            if verdict[3] != parent:
                return None, f"{row['id']}'s verdict line measured the base {verdict[3][:8]}, not the head's parent {str(parent)[:8]}"
            stale = fresh_base(repository, verdict[3], main)
            if stale:
                return None, f"{row['id']}, the newest verdict at the head, is stale: {stale}"
    return newest[0]["id"], "fresh and green"


def table_cell(text):
    """A table cell's value: its text, one pair of backticks around it taken off."""
    text = text.strip()
    return text[1:-1] if len(text) > 1 and text[0] == text[-1] == "`" else text


def declared_mutants(diff):
    """Each mutant the head's own *.mutants.md files declare, as a dict: its name (<file>:<row>), and, when its row is in
    the machine form, the patch, test and run cloud/gates/mutants-red.sh reads from it, else None for each and why.

    A row of a table headed `| Mutant |` is one mutant; a file with no such table is one mutant, run by no recipe. A row
    the recipe runs names, in columns headed Patch, Red and Run, an exact patch in the head's tree, the one test that
    goes red with it applied, and the command that runs that test. Two mutants restoring the same test are two."""
    declared = []
    for path, status in sorted(diff.status.items()):
        if status == "D" or not path.endswith(".mutants.md"):
            continue
        header, rows = None, 0
        for line in (diff.source(diff.head, path) or "").splitlines():
            cells = [cell.strip() for cell in line.strip().strip("|").split("|")] if line.lstrip().startswith("|") else None
            if cells is None:
                header = None
            elif cells[0].lower() == "mutant":
                header = [cell.lower() for cell in cells]
            elif header is not None and not set("".join(cells)) <= set("-: "):
                rows += 1
                declared.append(mutant_row(f"{path}:{rows}", dict(zip(header, (table_cell(cell) for cell in cells)))))
        if not rows:
            declared.append(dict(name=f"{path}:1", patch=None, test=None, run=None,
                                 why=f"{path} holds no `| Mutant |` table, so no recipe can run its mutant"))
    return declared


def mutant_row(name, cells):
    """One declared mutant from its row's cells, keyed by their header."""
    patch, test, run = cells.get("patch", ""), cells.get("red", ""), cells.get("run", "")
    unread = [column for column, cell, form in (("Patch", patch, MutantPatchCell), ("Red", test, MutantTestCell),
                                                ("Run", run, MutantRunCell)) if not form.match(cell)]
    if unread or not MutantName.match(name):
        return dict(name=name, patch=None, test=None, run=None,
                    why=f"{name} isn't in the machine form: its {', '.join(unread) or 'name'} cell doesn't read "
                        "(Patch an exact .patch path, Red one test, Run its command)")
    return dict(name=name, patch=patch, test=test, run=run, why=None)


def mutant_command(recipe, mutant):
    """The command the mutants-red recipe asks for one mutant: the recipe with the mutant's name written in."""
    if recipe.count(MutantPlaceholder) != 1:
        raise Unreadable(f"cloud/gates/mutants-red.sh holds {MutantPlaceholder} {recipe.count(MutantPlaceholder)} times, "
                         "not once")
    return recipe.replace(MutantPlaceholder, mutant["name"])


def mutant_verdict(mutant, head):
    """The line the mutants-red recipe ends on when the mutant's own test went red, and only then."""
    run = hashlib.sha256(mutant["run"].encode()).hexdigest()
    return f"VERDICT RED mutants-red {mutant['name']} patch {mutant['patch']} test {mutant['test']} run {run} head {head}"


def judge_mutant(mutant, job, recipe, rows, head):
    """Why the named job doesn't hold the mutant red at the head, or None when it does. Every check is exact: the job's
    command is this mutant's mutants-red.sh, at adamic and the head, the newest verdict of that command there, ended
    Exit with a code not 0, and its tail's last line is the verdict the recipe prints only when this mutant's own test
    failed, naming its patch, its test, its run and the head."""
    if mutant["test"] is None:
        return mutant["why"]
    row = next((row for row in rows if row["id"] == job), None)
    if row is None:
        return f"{job} isn't among the rows given"
    if row.get("commandSha256") != recipe_hash(mutant_command(recipe, mutant)):
        return f"{job}'s command isn't {mutant['name']}'s mutants-red.sh"
    if row.get("atRepository") != "adamic" or row.get("atSha") != head:
        return f"{job} ran at {row.get('atRepository')}@{str(row.get('atSha'))[:8]}, not the head"
    newest, why = newest_runs(rows, row["commandSha256"], head)
    if newest is None:
        return f"{job}: {why}"
    if job not in [run["id"] for run in newest]:
        return f"{job} isn't the newest verdict of {mutant['name']} at the head ({', '.join(run['id'] for run in newest)} is)"
    for run in newest:
        if run.get("outcome") != "Exit" or not isinstance(run.get("exit"), int) or run.get("exit") == 0:
            return f"{run['id']} ended {run.get('outcome')} {run.get('exit')}, not red"
        lines = (run.get("tailText") or "").rstrip().splitlines()
        if not lines or lines[-1].strip() != mutant_verdict(mutant, head):
            return f"{run['id']} doesn't end on the line that names {mutant['test']} red for {mutant['name']}"
    return None


def owed_by_hand(gate):
    """What the accepter checks for a gate or row the script can't prove."""
    return "owed by hand, the accepter checks: " + (ByHandGates.get(gate) or ByHandRows[gate])


def judge_mutants(rows, head, diff, recipe):
    """(the jobs or None, why) for mutants-red: for each mutant the head's *.mutants.md declares, the newest run at the
    head of its own command, red on its own test (judge_mutant). The jobs are found by each mutant's command hash, so a
    job counts for one mutant alone."""
    if recipe is None:
        return None, "cloud/gates/mutants-red.sh isn't at main's tip, so no job can be matched to a mutant"
    declared = declared_mutants(diff)
    if not declared:
        return None, "the head's diff holds no *.mutants.md declaring its mutants"
    held, why = [], []
    for mutant in declared:
        if mutant["test"] is None:
            why.append(mutant["why"])
            continue
        wanted = recipe_hash(mutant_command(recipe, mutant))
        newest, unread = newest_runs(rows, wanted, head)
        if newest is None:
            why.append(f"{mutant['name']}: {unread}")
            continue
        refused = judge_mutant(mutant, newest[0]["id"], recipe, rows, head)
        if refused:
            why.append(refused)
        else:
            held.append(f"{mutant['name']} {newest[0]['id']}")
    if why:
        return None, "; ".join(why)
    return ", ".join(held), f"{len(held)} jobs at the head, each its own mutant's recipe, red on its own test"


def added_skips(diff):
    """Each skip the diff adds in a test, as <path>:<test> (its enclosing Test function in the head), or <path>:<line>
    for a skip outside one."""
    skips = []
    for path, lines in sorted(diff.added_numbered.items()):
        if not fnmatch.fnmatch(path, "*_test.go"):
            continue
        source = (diff.source(diff.head, path) or "").splitlines()
        for number, line in lines:
            if not re.search(r"\bt\.Skip(f|Now)?\(", line):
                continue
            enclosing = next((found.group(1) for found in (re.match(r"func (Test\w+)\(", text)
                                                           for text in reversed(source[:number])) if found), None)
            skip = f"{path}:{enclosing or number}"
            if skip not in skips:
                skips.append(skip)
    return skips


def ask_line(gate, recipe, repository, head, main, task, gates_directory, mutant=None):
    """The exact ask for a missing gate. Its command is the recipe as main's tip holds it, read by git at main's sha, so
    neither the head nor a working tree can change it; with --gates, as that directory holds it. The job runs at the
    head. A mutant's ask writes its name into the recipe's placeholder, as mutant_command does, so the job's hash is the
    one the script holds it to."""
    flags = " ".join(shlex.quote(flag) for flag in ask_flags(recipe))
    if gates_directory is None:
        source = f"git -C {shlex.quote(str(repository.root))} show {main}:cloud/gates/{gate}.sh"
    else:
        source = f"cat {shlex.quote(str(gates_directory.resolve() / (gate + '.sh')))}"
    title = f"{gate} at {head[:8]}"
    if mutant is not None:
        source += f" | sed 's|{MutantPlaceholder}|{mutant['name']}|'"
        title = f"{gate} {mutant['name']} at {head[:8]}"
    return (f'ahra loom job "$({source})" --kind proof --task {task} --title "{title}" '
            f'--at adamic@{head}' + (f" {flags}" if flags else ""))


def choice_of(rule):
    """A rule's thenIf as a choice: (the identity gate, the value branch's gates, its rows). The map writes it as the
    identity gate's name set false beside what that owes; held true, the gate drops the branch (its description), so the
    head proves its JavaScript by the identity or by the whole branch."""
    then = rule["thenIf"]
    conditions = [key for key, value in then.items() if isinstance(value, bool)]
    return conditions[0], then.get("requires", []), then.get("sameDiff", [])


def merged_heads(repository, head, main):
    """Each merge among the head's own commits that joins two lines main doesn't hold, with those parents. A merge of
    main into the head joins one such line, so it is still one head."""
    joined = []
    for merge in repository.text("rev-list", "--merges", f"{main}..{head}").split():
        parents = repository.text("rev-list", "--parents", "-n", "1", merge).split()[1:]
        off = [parent for parent in parents if not repository.ancestor(parent, main)]
        if len(off) > 1:
            joined.append((merge, off))
    return joined


def judge(arguments):
    repository = Repository(arguments.repository)
    head, main = repository.commit(arguments.head), repository.commit(arguments.main)
    gates_directory = Path(arguments.gates) if arguments.gates else None

    def gate_file(name):
        if gates_directory is not None:
            path = gates_directory / name
            if not path.exists():
                return None
            with open(path, newline="") as file:
                return file.read()
        data = repository.show(main, "cloud/gates/" + name)
        return None if data is None else data.decode()

    # The judge runs as main holds it: this script's own bytes must be main's copy, where main has one
    held = repository.show(main, ScriptPath)
    if held is not None and held != Path(__file__).read_bytes():
        raise Unreadable(f"this {ScriptPath} isn't main's at {main[:8]}: the judge runs as main holds it, "
                         f"git -C {repository.root} show {main}:{ScriptPath}")
    map_text = gate_file("diff-to-gates.json")
    if map_text is None:
        raise Unreadable(f"main {main[:8]} carries no {MapPath}; a main cut before it is read with --gates <directory> "
                         "(the bootstrap, cloud/gates/README.md)")
    gates_map = read_map(map_text)
    diff = Diff(repository, head, main)
    diff.gates_map = gates_map
    rows = read_rows(arguments.jobs)

    fired = fire(gates_map, diff)
    owed, rows_owed, notes = {}, {}, []
    for rule in gates_map["rules"]:
        if rule["id"] not in fired:
            continue
        for gate in rule.get("requires", []):
            owed.setdefault(gate, rule["id"])
        for row in rule.get("sameDiff", []):
            rows_owed.setdefault(row, rule["id"])
        if "routing" in rule:
            notes.append(f"{rule['id']}: {rule['routing']}")
        if "accepter" in rule.get("thenIf", {}):
            notes.append(f"{rule['id']}: its accepter is {rule['thenIf']['accepter']}")
        if "fresh" in rule:
            notes.append(f"{rule['id']}: {rule['fresh']}; this script reports one head, so a set's combined census "
                         "(freshStack) is the accepter's to read")

    report = {"head": head, "main": main, "base": diff.base, "commits": len(diff.commits), "liveHole": arguments.live_hole,
              "fired": fired, "gates": [], "rows": [], "choices": [], "notes": notes}
    # A head is one change's diff against main's tip (Adamic, 14:33Z): a merge of main into it is still one head, and two
    # heads merged together are two, which the report says
    joined = merged_heads(repository, head, main)
    report["rows"].append({"name": "one-head", "rule": "a head is one change", "satisfied": not joined,
                           "why": f"one change on {diff.base[:8]}, {len(diff.commits)} commits" if not joined else "; ".join(
                               f"{merge[:8]} merges {len(parents)} heads main doesn't hold ({', '.join(p[:8] for p in parents)}), "
                               "so this is more than one head: report each on its own" for merge, parents in joined)})
    if arguments.live_hole:
        notes.append("--live-hole: the caller vouches that the task is a live hole on its two-hour clock and that its "
                     "pins ride the line; this script can see neither")
    if diff.base != main:
        notes.append(f"the head's base {diff.base[:8]} isn't main's tip {main[:8]}: a paired gate's base is held by fresh")
    if held is None:
        notes.append(f"bootstrap: main {main[:8]} holds no {ScriptPath}, so this run's judging code is the caller's, "
                     "never main's (cloud/gates/README.md)")
    if gates_directory is not None:
        notes.append(f"--gates: the map and recipes are {gates_directory}'s, never main's")
    # The head's own judging code: changed, a green of its recipes is the head judging itself, so it's owed by hand
    changed = [path for path in diff.paths() if path.startswith(JudgePaths)]
    report["rows"].append({"name": "gates-changed", "rule": "the judge's own code", "byHand": True, "satisfied": not changed,
                           "why": "the head changes no judging code" if not changed else
                           owed_by_hand("gates-changed") + f"; it changes {', '.join(changed)}"})

    def gate_entry(gate, rule, deferred=False):
        recipe = gate_file(gate + ".sh")
        entry = {"name": gate, "rule": rule, "deferred": deferred, "job": None, "ask": None}
        if gate in ByHandGates:
            entry["why"], entry["byHand"] = owed_by_hand(gate), True
        elif gate == "mutants-red":
            entry["job"], entry["why"] = judge_mutants(rows, head, diff, recipe)
            if entry["job"] is None and recipe is not None:
                entry["ask"] = [ask_line(gate, recipe, repository, head, main, arguments.task, gates_directory, mutant)
                                for mutant in declared_mutants(diff) if mutant["test"] is not None]
        elif recipe is None:
            entry["why"] = f"cloud/gates/{gate}.sh isn't at main's tip, so no job can be matched to it"
        else:
            entry["job"], entry["why"] = judge_gate(gate, recipe, rows, head, main, repository)
            if entry["job"] is None:
                entry["ask"] = ask_line(gate, recipe, repository, head, main, arguments.task, gates_directory)
        entry["satisfied"] = entry["job"] is not None
        return entry

    for gate, rule in owed.items():
        deferred = arguments.live_hole and gate not in LiveHoleGates
        report["gates"].append(gate_entry(gate, rule, deferred))
    if arguments.live_hole:
        for gate in LiveHoleGates:
            if gate not in owed:
                report["gates"].append(gate_entry(gate, "two-hour-clock"))
    # Each fired rule's thenIf, a choice: its identity gate held, or its whole value branch held. Unless the identity
    # holds, the branch is owed; either way the identity is no requirement alone, unless another fired rule requires it.
    for rule in gates_map["rules"]:
        if rule["id"] not in fired or "thenIf" not in rule:
            continue
        identity, value_gates, value_rows = choice_of(rule)
        if any(identity in other.get("requires", []) for other in gates_map["rules"]
               if other["id"] in fired and other["id"] != rule["id"]):
            continue
        entry = next(entry for entry in report["gates"] if entry["name"] == identity)
        if not entry["satisfied"]:
            for gate in value_gates:
                if gate not in owed:
                    report["gates"].append(gate_entry(gate, f"{rule['id']}, unless {identity}",
                                                      arguments.live_hole and gate not in LiveHoleGates))
            for row in value_rows:
                rows_owed.setdefault(row, f"{rule['id']}, unless {identity}")
        report["choices"].append({"rule": rule["id"], "identity": identity, "value": value_gates + value_rows})

    for row, rule in rows_owed.items():
        satisfied, why = RowReaders[row](diff, rule)
        entry = {"name": row, "rule": rule, "satisfied": satisfied, "why": why, "byHand": row in ByHandRows}
        report["rows"].append(entry)

    entries = {entry["name"]: entry for entry in report["gates"] + report["rows"]}
    for choice in report["choices"]:
        if entries[choice["identity"]]["satisfied"]:
            choice["by"] = "identity"
        elif all(entries[name]["satisfied"] for name in choice["value"]):
            choice["by"] = "value"
        else:
            choice["by"] = None
    # A choice held by either side holds every name in it; one held by neither leaves each unheld name owed.
    chosen = {name for choice in report["choices"] if choice["by"] for name in [choice["identity"], *choice["value"]]}
    report["missing"] = [entry["name"] for entry in report["gates"] + report["rows"]
                         if not entry["satisfied"] and not entry.get("deferred") and entry["name"] not in chosen]
    report["owedToday"] = [entry["name"] for entry in report["gates"]
                           if not entry["satisfied"] and entry.get("deferred") and entry["name"] not in chosen]
    return report


Advisory = ("advisory: this checklist never authorizes a submit; a gate shown as met is one the accepter still checks, "
            "and the accepter, Compiler or Adamic, is the authority")


def print_report(report):
    print(f"submit-gates: head {report['head'][:8]} on {report['base'][:8]}, main's tip {report['main'][:8]}"
          + (", a live hole on the two-hour clock" if report["liveHole"] else ""))
    print(Advisory)
    print("rules fired:")
    for rule, why in report["fired"].items():
        print(f"  {rule}: {why}")
    print("gates owed:")
    for entry in report["gates"]:
        state = "found" if entry["satisfied"] else ("owed today" if entry.get("deferred") else "unproven")
        print(f"  {state:10} {entry['name']:24} {entry['job'] or ''}  ({entry['rule']}) {entry['why']}".rstrip())
        if entry.get("ask") and not entry["satisfied"]:
            for ask in entry["ask"] if isinstance(entry["ask"], list) else [entry["ask"]]:
                print(f"             ask: {ask}")
    print("same diff:")
    for entry in report["rows"]:
        print(f"  {'found' if entry['satisfied'] else 'unproven':10} {entry['name']:24} ({entry['rule']}) {entry['why']}")
    for choice in report["choices"]:
        held = {"identity": f"evidence for {choice['identity']}", "value": "evidence for the value branch"}.get(
            choice["by"], "evidence for neither")
        print(f"choice ({choice['rule']}): {choice['identity']}, or {' and '.join(choice['value'])}: {held}")
    for note in report["notes"]:
        print(f"note: {note}")
    if report["owedToday"]:
        print(f"owed today, each run named on the live-hole task within the day: {', '.join(report['owedToday'])}")
    if report["missing"]:
        print(f"verdict: {len(report['missing'])} owed with no evidence found, each for the accepter to check: "
              f"{', '.join(report['missing'])}")
    else:
        print("verdict: evidence found for every owed gate and row, each still for the accepter to check")
    print(Advisory)


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--head", required=True, action="append", help="the head's sha; several are each reported alone")
    parser.add_argument("--main", required=True, help="main's tip at submit")
    parser.add_argument("--repository", required=True, help="an adamic clone holding both")
    parser.add_argument("--jobs", required=True, help="the Loom job rows, as `ahra loom job <id> --json` prints them")
    parser.add_argument("--task", required=True, help="the task each missing gate's ask names")
    parser.add_argument("--live-hole", action="store_true", help="the two-hour clock: submit on the live-hole gates alone")
    parser.add_argument("--gates", help="read the map and recipes from this directory, not main's tip")
    parser.add_argument("--json", action="store_true", help="print the report as JSON")
    arguments = parser.parse_args()
    heads = arguments.head
    if len(heads) > 1:
        print(f"submit-gates: {len(heads)} heads given, each reported on its own; a set's combined census (the map's "
              "freshStack) is the accepter's to read")
    reports = []
    for head in heads:
        arguments.head = head
        try:
            reports.append(judge(arguments))
        except (Unreadable, OSError, UnicodeError) as error:
            print(f"submit-gates: {error}", file=sys.stderr)
            return 2
    for report in reports:
        if arguments.json:
            print(json.dumps(report, indent=2))
        else:
            print_report(report)
    return 1 if any(report["missing"] for report in reports) else 0


if __name__ == "__main__":
    sys.exit(main())
