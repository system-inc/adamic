from pathlib import Path
import subprocess
import sys

base = Path(__file__).resolve().parent
source = base / "OctalEscape.a.txt"
manifest = base / "profile-invalid-corpus.txt"
manifest.write_text(str(source) + "\tno-var\t\t\tfalse\t\tnormal\n")
oracle = sys.argv[1] if len(sys.argv) > 1 else "/tmp/slot02-profile-b4691483/oracle"
evidence = base.parent / "evidence"
with (evidence / "profile-reproducer-diagnostics.log").open("wb") as log:
    diagnostics = subprocess.run([oracle, "--manifest", str(manifest), "--diagnostics"], stdout=log, stderr=subprocess.STDOUT)
with (evidence / "profile-reproducer.log").open("wb") as log:
    failure = subprocess.run([oracle, "--manifest", str(manifest)], stdout=log, stderr=subprocess.STDOUT)
(evidence / "profile-reproducer-exit.log").write_text(str(failure.returncode) + "\n")
assert diagnostics.returncode == 0
assert (evidence / "profile-reproducer-diagnostics.log").read_text() == "1\n"
assert failure.returncode == 2
assert "Octal escape sequences are not allowed" in (evidence / "profile-reproducer.log").read_text()
print("Expected shared failure reproduced: parse diagnostic 1, oracle exit 2 with octal-escape panic")
