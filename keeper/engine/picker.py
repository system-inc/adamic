"""picker.py: the standing mutant engine's first step (#ysavcpk, design #j3cbpzp section 1, @system_adamic_tests).

	picker.py <repo> <base> <head> <out dir> [--max 8]
	# writes <out dir>/<hash>.diff per mutant and <out dir>/manifest.json, then prints one line per mutant

The mutants come off the fixed menu, on lines the landed range base..head added or changed, and only in code under test:
Go (not _test.go), the C runtime under internal/native/runtime, and stage1 .ts and .a sources, each under a package that
has tests. Test files, testdata and generated files are never read, so no mutant is aimed at a test. At most --max per
range, taken round-robin across the changed files so one big file can't take them all, in an order seeded by head, so
the same range always yields the same mutants.

The menu, as the audit used it:
	flip-condition    a comparison in a condition becomes its opposite (== and !=, < and >=, > and <=), or && and || swap
	off-by-one        a comparison's boundary moves (< and <=, > and >=)
	change-constant   an integer literal n becomes n + 1
	drop-statement    an assignment, increment or call statement is removed
	swap-arguments    a call with exactly two plain arguments gets them in the other order
	return-early      not picked here: a correct early return needs the function's result types; the pool side can add it

A Go mutant that no longer parses (gofmt -e) is discarded and another taken. Whether it compiles is the replay's to say.
"""
import argparse, difflib, hashlib, json, os, random, re, shutil, subprocess, sys, tempfile

menu = ["flip-condition", "off-by-one", "change-constant", "drop-statement", "swap-arguments"]
flip = {"==": "!=", "!=": "==", "<": ">=", ">=": "<", ">": "<=", "<=": ">"}
boundary = {"<": "<=", "<=": "<", ">": ">=", ">=": ">"}


def git(repo, *arguments):
    return subprocess.run(["git", "-C", repo, *arguments], check=True, capture_output=True, text=True).stdout


def language(path):
    if path.endswith(".go"):
        return "go"
    if path.startswith("internal/native/runtime/") and path.endswith((".c", ".h")):
        return "c"
    if path.startswith("stage1/") and path.endswith((".ts", ".a")):
        return "ts"
    return None


tested = {}  # (repo, head) -> every directory holding a _test.go


def eligible(repo, head, path, text):
    """Code under test: the right kind of file, never a test or testdata or generated, under a package with tests."""
    if language(path) is None or path.endswith("_test.go") or "/testdata/" in "/" + path or path.startswith("review/"):
        return False
    if text.startswith("// Code generated") or "\n// Code generated" in text[:500]:
        return False
    if language(path) == "c":
        return True  # the runtime's home package, internal/native, has tests
    if (repo, head) not in tested:
        tested[(repo, head)] = {os.path.dirname(name) for name in git(repo, "ls-tree", "-r", "--name-only", head).splitlines() if name.endswith("_test.go")}
    directory = os.path.dirname(path)
    while True:  # a stage1 port's sources sit below the package that tests them
        if directory in tested[(repo, head)]:
            return True
        if language(path) == "go" or not directory:
            return False
        directory = os.path.dirname(directory)


def changed_lines(repo, base, head, path):
    """The new side's line numbers (1-based) that the range added or changed. Pure deletions leave nothing to mutate."""
    lines = set()
    for hunk in re.finditer(r"^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@", git(repo, "diff", "-U0", base, head, "--", path), re.M):
        start, count = int(hunk.group(1)), int(hunk.group(2) or "1")
        lines.update(range(start, start + count))
    return lines


def mask(line):
    """The line with string, rune and template literal contents and trailing // comments blanked to spaces, so a
    candidate is never found inside them. Same length, so offsets stay true."""
    masked, quote, index = list(line), None, 0
    while index < len(line):
        character = line[index]
        if quote:
            if character == "\\" and quote != "`":
                masked[index] = " "
                if index + 1 < len(line):
                    masked[index + 1] = " "
                index += 2
                continue
            if character == quote:
                quote = None
            else:
                masked[index] = " "
        elif character in "\"'`":
            quote = character
        elif line.startswith("//", index):
            for rest in range(index, len(line)):
                masked[rest] = " "
            break
        index += 1
    return "".join(masked)


comparison = re.compile(r"(?<![<>=!&|])(==|!=|<=|>=|<|>)(?![<>=])")


def candidates(path, lines, number):
    """Every menu mutant for one line, as (op, new line)."""
    line, kind = lines[number - 1], language(path)
    masked, stripped = mask(line), line.strip()
    found = []
    if not stripped or stripped.startswith(("//", "/*", "*", "import", "package", "#")):
        return found
    conditional = re.match(r"\s*(\}\s*else\s+)?(if|for|while)\b", masked) or re.search(r"\breturn\b.*(==|!=|&&|\|\|)", masked)
    if conditional:
        for match in comparison.finditer(masked):
            operator = match.group(1)
            if kind == "go" and operator in "<>" and re.search(r"\w\[?$", masked[:match.start()]) and masked[match.end():match.end() + 1] in "]":
                continue  # a generic's brackets, not a comparison
            found.append(("flip-condition", line[:match.start()] + flip[operator] + line[match.end():]))
            if operator in boundary:
                found.append(("off-by-one", line[:match.start()] + boundary[operator] + line[match.end():]))
        for match in re.finditer(r"&&|\|\|", masked):
            found.append(("flip-condition", line[:match.start()] + ("||" if match.group() == "&&" else "&&") + line[match.end():]))
    for match in re.finditer(r"(?<![\w.])(\d+)(?![\w.])", masked):
        if re.search(r"\[\s*$", masked[:match.start()]) and kind == "go" and re.match(r"\s*\]\s*\w", masked[match.end():]):
            continue  # an array type's length: changing it changes the type, not the behavior
        found.append(("change-constant", line[:match.start()] + str(int(match.group(1)) + 1) + line[match.end():]))
    statement = re.match(r"\s*([\w.\[\]*]+\s*([+\-*/|&]?=|\+\+|--)|[\w.]+\(.*\)\s*;?\s*$)", masked)
    if statement and ":=" not in masked and not re.match(r"\s*(var|const|type|func|let|return|if|for|switch|case|defer|go)\b", masked):
        if not masked.rstrip().endswith(("{", "(", ",")):
            found.append(("drop-statement", None))
    for match in re.finditer(r"\b[\w.]+\(\s*([\w.&*]+)\s*,\s*([\w.&*]+)\s*\)", masked):
        first, second = match.group(1), match.group(2)
        if first != second:
            start = match.start(1)
            found.append(("swap-arguments", line[:start] + second + ", " + first + line[match.end(2):]))
    return found


def parses(kind, text):
    if kind != "go" or shutil.which("gofmt") is None:
        return True
    with tempfile.NamedTemporaryFile("w", suffix=".go", delete=False) as handle:
        handle.write(text)
    try:
        return subprocess.run(["gofmt", "-e", "-l", handle.name], capture_output=True).returncode == 0
    finally:
        os.unlink(handle.name)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("repo")
    parser.add_argument("base")
    parser.add_argument("head")
    parser.add_argument("out")
    parser.add_argument("--max", type=int, default=8)
    options = parser.parse_args()
    head = git(options.repo, "rev-parse", options.head).strip()
    random.seed(head)
    pools = []  # per file: shuffled (path, number, op, new line, original text)
    for path in sorted(git(options.repo, "diff", "--name-only", "--diff-filter=AM", options.base, head).splitlines()):
        if language(path) is None:
            continue
        text = git(options.repo, "show", head + ":" + path)
        if not eligible(options.repo, head, path, text):
            continue
        lines = text.split("\n")
        pool = [(path, number, op, new, text) for number in sorted(changed_lines(options.repo, options.base, head, path)) if number <= len(lines) for op, new in candidates(path, lines, number)]
        random.shuffle(pool)
        if pool:
            pools.append(pool)
    os.makedirs(options.out, exist_ok=True)
    chosen, seen = [], set()
    while len(chosen) < options.max and any(pools):
        for pool in pools:
            while pool and len(chosen) < options.max:
                path, number, op, new, text = pool.pop()
                lines = text.split("\n")
                mutated = lines[:number - 1] + ([] if new is None else [new]) + lines[number:]
                after = "\n".join(mutated)
                if after == text or (path, number, after) in seen or not parses(language(path), after):
                    continue
                seen.add((path, number, after))
                diff = "".join(difflib.unified_diff([line + "\n" for line in lines], [line + "\n" for line in mutated], "a/" + path, "b/" + path))
                hash = hashlib.sha256(diff.encode()).hexdigest()[:12]
                open(os.path.join(options.out, hash + ".diff"), "w").write(diff)
                chosen.append({"hash": hash, "file": path, "line": number, "op": op, "before": lines[number - 1].strip(), "after": "" if new is None else new.strip()})
                break
    json.dump({"base": git(options.repo, "rev-parse", options.base).strip(), "head": head, "menu": menu, "mutants": chosen}, open(os.path.join(options.out, "manifest.json"), "w"), indent=1)
    for mutant in chosen:
        print("%s\t%s:%d\t%s" % (mutant["hash"], mutant["file"], mutant["line"], mutant["op"]))


if __name__ == "__main__":
    main()
