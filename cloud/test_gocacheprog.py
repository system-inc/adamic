#!/usr/bin/env python3
"""cloud/gocacheprog.sh on a fake Go and a fake curl: the pin, the env.sh block, an older env.sh; no network."""
import io
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import unittest

SOURCE = Path(os.environ.get("ADAMIC_GOCACHEPROG_SOURCE", Path(__file__).resolve().parent))
PIN = next(line.split("=", 1)[1] for line in (SOURCE / "go-pin.sh").read_text().splitlines()
           if line.startswith("goPin="))
OTHER = "go1.27.2" if PIN != "go1.27.2" else "go1.27.1"

FAKE_GO = """#!/bin/sh
case "$1" in
	env) [ "$2" = GOVERSION ] && echo "{version}" ;;
	version) echo "go version {version} linux/amd64" ;;
	build) while [ $# -gt 0 ]; do [ "$1" = -o ] && out=$2; shift; done; printf '#!/bin/sh\\n' > "$out"; chmod +x "$out" ;;
esac
"""
FAKE_CURL = """#!/bin/sh
[ -n "${FAKE_GO_TARBALL:-}" ] || exit 22
cat "$FAKE_GO_TARBALL"
"""
# What setup.sh wrote before cloud/gocacheprog.sh existed (main 17:21Z to this change).
LEGACY = """export PATH="{tools}/bin:{tools}/go/bin:$PATH"
export GOTOOLCHAIN=auto
export TMPDIR=/tmp/adamic-gate
# A product's bytes are a function of its key (@system_adamic, Oct 9): without these, every binary stamps the commit
# and the checkout path, inputs no key sees. Added to whatever GOFLAGS holds, once, however often this is sourced.
case " $GOFLAGS " in *" -buildvcs=false -trimpath "*) ;; *) export GOFLAGS="${{GOFLAGS:+$GOFLAGS }}-buildvcs=false -trimpath" ;; esac
unset GOCACHEPROG
if [ "${{ADAMIC_GOCACHE_OFF:-0}}" != 1 ] && [ -d {repository}/cmd/adamic-gocacheprog ] && [ -x {tools}/bin/old-program ]; then
	export GOCACHEPROG={tools}/bin/old-program
fi
export WASI_SYSROOT=/opt/wasi
"""


class GoCacheProg(unittest.TestCase):
    def setUp(self):
        self.scratch = Path(tempfile.mkdtemp(prefix="gocacheprog-proof-"))
        self.addCleanup(shutil.rmtree, self.scratch, ignore_errors=True)
        self.repository = self.scratch / "repository"
        (self.repository / "cloud").mkdir(parents=True)
        for name in ["go-pin.sh", "gocacheprog.sh"]:
            shutil.copyfile(SOURCE / name, self.repository / "cloud" / name)
        (self.repository / "cmd/adamic-gocacheprog").mkdir(parents=True)
        (self.repository / "cmd/adamic-gocacheprog/main.go").write_text("package main\nfunc main() {}\n")
        self.tools = self.scratch / "tools"
        (self.tools / "bin").mkdir(parents=True)
        self.fakes = self.scratch / "fakes"
        self.fakes.mkdir()
        self.executable(self.fakes / "curl", FAKE_CURL)
        self.environment = dict(os.environ, PATH=f"{self.fakes}:{os.environ['PATH']}")
        for name in ["GOCACHEPROG", "GOFLAGS", "ADAMIC_GOCACHE_OFF", "FAKE_GO_TARBALL"]:
            self.environment.pop(name, None)
        (self.tools / "env.sh").write_text(f'export PATH="{self.tools}/bin:{self.tools}/go/bin:$PATH"\n'
                                           "export TMPDIR=/tmp/adamic-gate\n")

    def executable(self, path, text):
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)
        path.chmod(0o755)

    def go(self, version):
        self.executable(self.tools / "go/bin/go", FAKE_GO.format(version=version))

    def tarball(self, version):
        archive = self.scratch / f"{version}.tar.gz"
        with tarfile.open(archive, "w:gz") as tar:
            data = FAKE_GO.format(version=version).encode()
            entry = tarfile.TarInfo("go/bin/go")
            entry.size, entry.mode = len(data), 0o755
            tar.addfile(entry, io.BytesIO(data))
        return archive

    def run_script(self, **extra):
        result = subprocess.run(["bash", str(self.repository / "cloud/gocacheprog.sh"), str(self.tools)],
                                env=dict(self.environment, **extra), capture_output=True, text=True, timeout=60)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result.stdout

    def sourced(self):
        result = subprocess.run(["bash", "-c", f'source {self.tools}/env.sh && echo "$GOCACHEPROG|$GOFLAGS|$PATH"'],
                                env=self.environment, capture_output=True, text=True, check=True)
        program, flags, path = result.stdout.strip().split("|", 2)
        return program, flags, path

    def test_pinned_go_sets_the_program(self):
        self.go(PIN)
        self.assertIn("shared cache ready", self.run_script())
        program, flags, path = self.sourced()
        self.assertEqual(program, str(self.tools / "bin/adamic-gocacheprog"))
        self.assertTrue(os.access(program, os.X_OK))
        self.assertIn("-buildvcs=false -trimpath", flags)
        self.assertIn(f"{self.tools}/go/bin", path.split(":"))

    def test_rerun_is_a_no_op(self):
        self.go(PIN)
        self.run_script()
        before = (self.tools / "env.sh").read_bytes()
        self.assertIn("unchanged", self.run_script())
        self.assertEqual(before, (self.tools / "env.sh").read_bytes())

    def test_newer_legacy_flags_line_is_replaced(self):
        self.go(PIN)
        (self.tools / "env.sh").write_text(LEGACY.format(tools=self.tools, repository=self.repository)
                                           .replace('case " $GOFLAGS "', 'case " ${GOFLAGS:-} "'))
        self.run_script()
        text = (self.tools / "env.sh").read_text()
        self.assertEqual(text.count('case " ${GOFLAGS:-} "'), 1, text)

    def test_sourced_under_set_u(self):
        self.go(PIN)
        self.run_script()
        environment = dict(self.environment)
        environment.pop("GOFLAGS", None)
        result = subprocess.run(["bash", "-uc", f"source {self.tools}/env.sh && echo \"$GOFLAGS\""],
                                env=environment, capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("-trimpath", result.stdout)

    def test_changed_source_rebuilds(self):
        self.go(PIN)
        self.run_script()
        (self.repository / "cmd/adamic-gocacheprog/main.go").write_text("package main\nfunc main() { _ = 1 }\n")
        output = self.run_script()
        self.assertNotIn("unchanged", output)
        self.assertEqual((self.tools / "env.sh").read_text().count("# gocacheprog end"), 1)

    def test_older_env_is_replaced_not_stacked(self):
        self.go(PIN)
        (self.tools / "env.sh").write_text(LEGACY.format(tools=self.tools, repository=self.repository))
        self.run_script()
        text = (self.tools / "env.sh").read_text()
        self.assertEqual(text.count("export GOCACHEPROG="), 1, text)
        self.assertEqual(text.count("unset GOCACHEPROG"), 1, text)
        self.assertEqual(text.count('case " ${GOFLAGS:-} "'), 1, text)
        self.assertNotIn('case " $GOFLAGS "', text)
        self.assertNotIn("old-program", text)
        self.assertIn("export TMPDIR=/tmp/adamic-gate", text)
        self.assertIn("export WASI_SYSROOT=/opt/wasi", text)
        self.assertEqual(self.sourced()[0], str(self.tools / "bin/adamic-gocacheprog"))

    def test_another_go_is_replaced_by_the_pin(self):
        self.go(OTHER)
        output = self.run_script(FAKE_GO_TARBALL=str(self.tarball(PIN)))
        self.assertIn("shared cache ready", output)
        version = subprocess.run([str(self.tools / "go/bin/go"), "env", "GOVERSION"], capture_output=True, text=True)
        self.assertEqual(version.stdout.strip(), PIN)
        self.assertEqual(self.sourced()[0], str(self.tools / "bin/adamic-gocacheprog"))

    def test_split_toolchain_never_gets_the_program(self):
        self.go(OTHER)
        output = self.run_script()
        self.assertIn(f"go is {OTHER}, the pin is {PIN}", output)
        self.assertEqual(self.sourced()[0], "")
        # A failed download keeps the Go the machine had.
        version = subprocess.run([str(self.tools / "go/bin/go"), "env", "GOVERSION"], capture_output=True, text=True)
        self.assertEqual(version.stdout.strip(), OTHER)

    def test_off_switch(self):
        self.go(PIN)
        self.assertIn("ADAMIC_GOCACHE_OFF=1", self.run_script(ADAMIC_GOCACHE_OFF="1"))
        self.assertEqual(self.sourced()[0], "")


if __name__ == "__main__":
    unittest.main()
