#!/usr/bin/env python3
"""write-env.sh rewrites setup's own env.sh lines and keeps the ADAMIC_* exports another step wrote (#nwsyj7z)."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parent
WRITER = Path(os.environ.get("ADAMIC_WRITE_ENV", SOURCE / "write-env.sh"))

# The 20 gate-input exports a gate box's env.sh held when setup.sh's full rewrite dropped them (Server, 22:15Z Oct 9).
GATE_INPUTS = [
    "ADAMIC_TYPESCRIPT_SOURCE", "ADAMIC_CYCLE_LEDGER_ROOT", "ADAMIC_CYCLE_LEDGER_OUTPUT", "ADAMIC_CSS_FIXTURES",
    "ADAMIC_CSSNUMBERS_LIBRARY", "ADAMIC_CSSSTRINGS_LIBRARY", "ADAMIC_MARKDOWNINLINE_LIBRARY", "ADAMIC_CSS_LIBRARY",
    "ADAMIC_GRAPHQL_LIBRARY", "ADAMIC_MEDIA_QUERY_LIBRARY", "ADAMIC_SELECTOR_LIBRARY", "ADAMIC_VALUES_LIBRARY",
    "ADAMIC_GRAPHQL_PRETTIER", "ADAMIC_JSON_PRETTIER", "ADAMIC_CSS_PRINTER_LIBRARY", "ADAMIC_ESTREE_LIBRARY",
    "ADAMIC_TS_PRETTIER", "ADAMIC_YAML_LIBRARY", "ADAMIC_GITIGNORE_LARGEST", "ADAMIC_CLANG_TSGO_ARCHIVE",
]


class WriteEnvironment(unittest.TestCase):
    def setUp(self):
        self.tools = Path(tempfile.mkdtemp(prefix="write-env-"))

    def write(self, gocacheprog="on", wasi="/opt/adamic-tools/wasi-sdk"):
        subprocess.run(["bash", str(WRITER), str(self.tools), "/tmp/adamic-gate", str(self.tools / "markdown-width"),
                        "/repository", gocacheprog, wasi], check=True)
        return (self.tools / "env.sh").read_text()

    def test_a_box_keeps_its_gate_inputs_through_a_rewrite(self):
        gate_lines = [f"export {name}=/home/ahra/adamic-tools/gate-inputs/{name.lower()}" for name in GATE_INPUTS]
        (self.tools / "env.sh").write_text("export PATH=\"/old/bin:$PATH\"\nexport ADAMIC_MARKDOWNWIDTH_DEPS=\"/old\"\n"
                                           "unset GOCACHEPROG\n" + "\n".join(gate_lines) + "\n")
        written = self.write()
        lines = written.splitlines()
        for line in gate_lines:
            self.assertEqual(lines.count(line), 1, line)
        # setup's own lines are its own: the old PATH and markdown path are gone, the new ones written once.
        self.assertNotIn("/old/bin", written)
        self.assertEqual(sum(line.startswith("export ADAMIC_MARKDOWNWIDTH_DEPS=") for line in lines), 1)
        self.assertIn(f'export ADAMIC_MARKDOWNWIDTH_DEPS="{self.tools / "markdown-width"}"', lines)
        self.assertIn("export WASI_SYSROOT=/opt/adamic-tools/wasi-sdk/share/wasi-sysroot", lines)
        self.assertIn("\texport GOCACHEPROG=" + str(self.tools / "bin/adamic-gocacheprog"), lines)
        # Rewriting again changes nothing.
        self.assertEqual(self.write(), written)

    def test_a_fresh_instance_gets_only_setups_lines(self):
        written = self.write(gocacheprog="off", wasi="")
        self.assertNotIn("ADAMIC_TYPESCRIPT_SOURCE", written)
        self.assertNotIn("WASI_SYSROOT", written)
        self.assertNotIn("export GOCACHEPROG", written)
        self.assertTrue(written.startswith('export PATH="' + str(self.tools / "bin") + ':'))
        self.assertIn("unset GOCACHEPROG", written)

    def test_the_file_is_replaced_whole(self):
        self.write()
        self.assertFalse((self.tools / "env.sh.writing").exists())


if __name__ == "__main__":
    unittest.main()
