from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / "internal/lower/object.go"
witness = root / "internal/lower/fx5_mutant_witness_test.go"
green = source.read_text()
if witness.exists():
    raise RuntimeError("temporary witness already exists")
for name, fixture, refusal, catcher in [
    ("every", "every", "Every", "Native backend stdout"),
    ("findLast", "find_last", "FindLast", "from incompatible type 'adamic_heap *'"),
]:
    try:
        patch = Path(__file__).parent / ("revert-" + name + ".diff")
        applied = subprocess.run(["git", "apply", str(patch)], cwd=root, capture_output=True, text=True)
        if applied.returncode:
            raise RuntimeError(applied.stderr)
        witness.write_text('package lower\nimport ( "testing"; "github.com/system-inc/adamic/internal/native" )\nfunc TestFx5MutantWitness(t *testing.T) {\n t.Parallel()\n source := arrayNarrowingSource(t, "' + fixture + '")\n program := lowersAndAgreesWithNode(t, source)\n got := runAgreementNative(t, native.C(program))\n t.Logf("native stdout=%q stderr=%q exit=%d", got.stdout, got.stderr, got.code)\n want := runAgreementNode(t, "testdata/array_narrowing/' + fixture + '.a")\n t.Logf("Node stdout=%q stderr=%q exit=%d", want.stdout, want.stderr, want.code)\n compareNativeAgreement(t, got, want)\n}\n')
        command = ["go", "test", "./internal/lower", "-run", "^Test(ArrayNarrowing" + refusal + "Refused|Fx5MutantWitness)$", "-count=1", "-v", "-timeout", "90s"]
        log = Path(__file__).parent / ("mutant-" + name + ".log")
        with log.open("w") as output:
            output.write("command: " + " ".join(command) + "\n")
            output.flush()
            result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=180)
        text = log.read_text()
        if result.returncode == 0 or catcher not in text or "got <nil>" not in text:
            raise RuntimeError("mutant was not caught as expected: " + text)
        print(name + ": refusal test failed on acceptance; witness caught by " + catcher)
    finally:
        source.write_text(green)
        witness.unlink(missing_ok=True)
