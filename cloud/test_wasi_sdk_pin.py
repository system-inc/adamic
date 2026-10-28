#!/usr/bin/env python3
"""setup.sh's --wasi-sdk lines choose the SDK archive by the toolchain's clang, never `uname -m` (#9nqg4ex)."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parent


class WasiSDKPin(unittest.TestCase):
    pins = {"x86_64": "b7d4d944c88503e4f21d84af07ac293e3440b1b6210bfd7fe78e0afd92c23bc2",
            "arm64": "4cf4c553c4640e63e780442146f87d83fdff5737f988c06a6e3b2f0228e37665"}

    def choose(self, kernel, target):
        # The lines setup.sh runs from wasiVersion to wasiDirectory, cut from the file itself, so a pin chosen in
        # setup.sh and not by the script is read too.
        lines = (SOURCE / "setup.sh").read_text().splitlines()
        start = next(index for index, line in enumerate(lines) if line.strip().startswith("wasiVersion="))
        end = next(index for index in range(start, len(lines)) if lines[index].strip().startswith("wasiDirectory="))
        with tempfile.TemporaryDirectory(prefix="setup-wasi-pin-") as directory:
            scratch = Path(directory)
            (scratch / "cloud").mkdir()
            shutil.copyfile(SOURCE / "wasi-sdk-pin.sh", scratch / "cloud/wasi-sdk-pin.sh")
            (scratch / "tools/bin").mkdir(parents=True)
            (scratch / "fake").mkdir()
            for path, answer in [(scratch / "tools/bin/clang", target), (scratch / "fake/uname", kernel)]:
                path.write_text(f"#!/bin/sh\necho {answer}\n")
                path.chmod(0o755)
            script = "\n".join(["set -euo pipefail", f"repository={scratch}", f"tools={scratch / 'tools'}",
                                *lines[start:end + 1], 'echo "$wasiArchitecture $wasiSHA256"'])
            result = subprocess.run(["bash", "-c", script], capture_output=True, text=True, timeout=30,
                                    env=dict(os.environ, PATH=f"{scratch / 'fake'}:{os.environ['PATH']}"))
        return result

    def test_amd64_toolchain_on_an_arm64_kernel_lays_the_x86_64_sdk(self):
        result = self.choose("aarch64", "x86_64-unknown-linux-gnu")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.split(), ["x86_64", self.pins["x86_64"]])

    def test_arm64_toolchain_on_an_amd64_kernel_lays_the_arm64_sdk(self):
        result = self.choose("x86_64", "aarch64-unknown-linux-gnu")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.split(), ["arm64", self.pins["arm64"]])

    def test_a_target_with_no_sdk_fails_setup(self):
        result = self.choose("x86_64", "riscv64-unknown-linux-gnu")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("no WASI SDK", result.stderr)


if __name__ == "__main__":
    unittest.main()
