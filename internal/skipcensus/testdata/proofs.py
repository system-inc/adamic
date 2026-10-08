"""Run TestCensus against real source-copy annotation mutants."""
import os
import pathlib
import shutil
import subprocess
import tempfile

repo = pathlib.Path(__file__).resolve().parents[3]
root = pathlib.Path(tempfile.mkdtemp(prefix="census-annotations-mutants-"))
for name in subprocess.check_output(["git", "ls-files", "*_test.go"], cwd=repo, text=True).splitlines():
    target = root / name
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(repo / name, target)
for path in (repo / "internal/skipcensus").rglob("*_test.go"):
    target = root / path.relative_to(repo)
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(path, target)
binary = root / "census.test"
with (root / "build.log").open("w") as log:
    subprocess.run(["go", "test", "-c", "-o", str(binary), "./internal/skipcensus"], cwd=repo, stdout=log, stderr=subprocess.STDOUT, check=True)
for name, annotation, expected in [
    ("missing", "", "missing census annotation"),
    ("bad-class", "// census: optional unknown", "unknown census class optional"),
    ("no-variable", "// census: required-input setup supplies corpus", "must name its variable"),
]:
    target = root / "internal/skipcensus/new_skip_test.go"
    target.write_text('package skipcensus\nimport "testing"\nfunc TestAnnotationWitness(t *testing.T) {\n' + annotation + '\n t.Skip("missing")\n}\n')
    logfile = root / (name + ".log")
    with logfile.open("w") as log:
        result = subprocess.run([str(binary), "-test.run=^TestCensus$", "-test.v"], cwd=repo / "internal/skipcensus", env=dict(os.environ, ADAMIC_SKIP_CENSUS_ROOT=str(root)), stdout=log, stderr=subprocess.STDOUT)
    output = logfile.read_text()
    assert result.returncode == 1 and expected in output and "new_skip_test.go:5 (TestAnnotationWitness)" in output, output
    print(name + ": TestCensus exit 1, named file:5 and TestAnnotationWitness; " + str(logfile))
