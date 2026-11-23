source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin mutants-red || exit 1
# ask: --size F16 --ceiling 60m
# mutants-red: one mutant the head declares, applied at the head's own sha and run, red only on its own test
# (cloud/gates/diff-to-gates.json's red: exit non-zero, with the named test's FAIL line in its kept tail). The mutant is
# named on the line below by its declaration, <file>.mutants.md:<row>, its rows counted from 1 across the file's tables
# headed `| Mutant |`; cloud/submit-gates.py writes the name in where the placeholder stands when it asks, so each
# mutant's command, and its sha256, is its own. The row names, in its columns Patch, Red and Run, an exact patch in the
# head's tree, the one test the patch turns red, and the command that runs that test.
# The job ends on this line, exit 1, only when the run exited non-zero and its output holds that test's own FAIL line,
# Go's `--- FAIL: <test> (` or unittest's `FAIL: <test> (`:
#
#   VERDICT RED mutants-red <mutant> patch <patch> test <test> run <sha256 of the run> head <sha>
#
# A setup that fails, a run that passes, or another test's failure ends on gate_end's VERDICT FAIL, which holds nothing.
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gateMutant='@MUTANT@'
# The mutant's row, read from the head's own declaration: its patch, its test, its run's sha256 and its run, a line each.
python3 -I - "$gateMutant" > "$gateScratch/mutant" << 'PYTHON' || gate_abort declaration "the declaration of $gateMutant doesn't read"
import hashlib, re, subprocess, sys
name = sys.argv[1]
path, _, ordinal = name.rpartition(":")
if not re.fullmatch(r"[A-Za-z0-9._/-]+\.mutants\.md", path) or not re.fullmatch(r"[1-9][0-9]*", ordinal):
    sys.exit(f"the mutant name {name!r} does not read <file>.mutants.md:<row>")
shown = subprocess.run(["git", "show", "HEAD:" + path], capture_output=True)
if shown.returncode != 0:
    sys.exit(f"{path} is not in the head")
def value(text):
    text = text.strip()
    return text[1:-1] if len(text) > 1 and text[0] == text[-1] == "`" else text
header, rows = None, []
for line in shown.stdout.decode(errors="replace").splitlines():
    cells = [cell.strip() for cell in line.strip().strip("|").split("|")] if line.lstrip().startswith("|") else None
    if cells is None:
        header = None
    elif cells[0].lower() == "mutant":
        header = [cell.lower() for cell in cells]
    elif header is not None and not set("".join(cells)) <= set("-: "):
        rows.append(dict(zip(header, (value(text) for text in cells))))
if int(ordinal) > len(rows):
    sys.exit(f"{path} declares {len(rows)} mutants, not {ordinal}")
row = rows[int(ordinal) - 1]
patch, test, run = row.get("patch", ""), row.get("red", ""), row.get("run", "")
if not (re.fullmatch(r"[A-Za-z0-9._/-]+\.patch", patch) and re.fullmatch(r"[A-Za-z_][A-Za-z0-9_/.-]*", test)
        and re.fullmatch(r"[^`\n]+", run)):
    sys.exit(f"{name} is not in the machine form: Patch an exact .patch path, Red one test, Run its command")
print(patch, test, hashlib.sha256(run.encode()).hexdigest(), run, sep="\n")
PYTHON
{ read -r mutantPatch; read -r mutantTest; read -r mutantRunSha256; IFS= read -r mutantRun; } < "$gateScratch/mutant"
printf 'mutant %s: patch %s, test %s, run %s\n' "$gateMutant" "$mutantPatch" "$mutantTest" "$mutantRun"
if git apply --check "$mutantPatch" && git apply "$mutantPatch"; then
	gate_pass "applied $mutantPatch"
else
	gate_abort patch "$mutantPatch doesn't apply to the head"
fi
mutantExit=0
bash -c "$mutantRun" > "$gateScratch/run.log" 2>&1 || mutantExit=$?
printf 'run: exit=%s\n' "$mutantExit"
grep -E '^[[:space:]]*(--- )?FAIL' "$gateScratch/run.log" | head -20
if python3 -I - "$gateScratch/run.log" "$mutantTest" "$mutantExit" << 'PYTHON'
import sys
log, test, exit = sys.argv[1], sys.argv[2], int(sys.argv[3])
lines = open(log, errors="replace").read().splitlines()
red = any(line.strip().startswith(f"--- FAIL: {test} (") or line.startswith(f"FAIL: {test} (") for line in lines)
sys.exit(0 if exit != 0 and red else 1)
PYTHON
then
	printf 'VERDICT RED mutants-red %s patch %s test %s run %s head %s\n' "$gateMutant" "$mutantPatch" "$mutantTest" \
		"$mutantRunSha256" "$gateHead"
	exit 1
fi
tail -30 "$gateScratch/run.log"
gate_fail red "$mutantTest isn't red with $mutantPatch applied (the run exited $mutantExit)"
gate_end
