#!/usr/bin/env python3
"""source-env.sh makes a login shell that isn't interactive, Codex's `bash -lc`, find setup's tools (#ybkwr91)."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parent
WRITER = Path(os.environ.get("ADAMIC_SOURCE_ENV", SOURCE / "source-env.sh"))

# Debian's skeleton ~/.bashrc returns at once in a shell that isn't interactive, and its ~/.profile includes it.
BASHRC = "case $- in\n    *i*) ;;\n      *) return;;\nesac\nHISTCONTROL=ignoreboth\n"
PROFILE = 'if [ -n "$BASH_VERSION" ]; then\n    if [ -f "$HOME/.bashrc" ]; then\n\t. "$HOME/.bashrc"\n    fi\nfi\n'


class SourceEnvironment(unittest.TestCase):
    def setUp(self):
        self.home = Path(tempfile.mkdtemp(prefix="source-env-"))
        self.tools = self.home / "adamic-tools"
        (self.tools / "bin").mkdir(parents=True)
        program = self.tools / "bin/adamic-probe-tool"
        program.write_text("#!/bin/sh\necho found\n")
        program.chmod(0o755)
        (self.tools / "env.sh").write_text(f'export PATH="{self.tools / "bin"}:$PATH"\n')
        (self.home / ".bashrc").write_text(BASHRC)
        (self.home / ".profile").write_text(PROFILE)

    def write(self):
        subprocess.run(["bash", str(WRITER), str(self.tools / "env.sh"), str(self.home)], check=True)

    def shell(self, flags):
        environment = {"HOME": str(self.home), "PATH": "/usr/bin:/bin"}
        result = subprocess.run(["bash", flags, "command -v adamic-probe-tool"], env=environment,
                                stdin=subprocess.DEVNULL, capture_output=True, text=True)
        return result.stdout.strip()

    def test_a_login_shell_that_isnt_interactive_finds_the_tools(self):
        self.assertEqual(self.shell("-lc"), "", "the probe tool is found before setup wrote anything")
        self.write()
        self.assertEqual(self.shell("-lc"), str(self.tools / "bin/adamic-probe-tool"))

    def test_each_file_holds_the_line_once_however_often_setup_runs(self):
        self.write()
        self.write()
        line = f"source {self.tools / 'env.sh'}"
        for name in (".bashrc", ".profile"):
            self.assertEqual((self.home / name).read_text().splitlines().count(line), 1, name)

    def test_a_bash_profile_is_the_file_a_login_shell_reads(self):
        (self.home / ".bash_profile").write_text("")
        self.write()
        self.assertIn(f"source {self.tools / 'env.sh'}", (self.home / ".bash_profile").read_text())
        self.assertNotIn("env.sh", (self.home / ".profile").read_text())
        self.assertEqual(self.shell("-lc"), str(self.tools / "bin/adamic-probe-tool"))


if __name__ == "__main__":
    unittest.main()
