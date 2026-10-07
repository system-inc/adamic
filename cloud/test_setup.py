#!/usr/bin/env python3
"""Exercise warming keys and real setup invalidation; all subprocess output goes to logs."""
import copy
import importlib.util
import os
from pathlib import Path
import shutil
import sys
import subprocess
import tempfile
import unittest

sys.dont_write_bytecode = True
SOURCE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location(
    "setup_key", os.environ.get("ADAMIC_SETUP_KEY_MODULE", SOURCE / "setup-key.py")
)
key = importlib.util.module_from_spec(spec)
spec.loader.exec_module(key)


class WarmingKey(unittest.TestCase):
    def test_every_component_invalidates(self):
        inputs = dict(head="commit", sums=b"sums", version="go1.27.1",
                      environment={"GOFLAGS": "", "setup-source": "script"},
                      packages=[["package", "build-id"]], cache=[["artifact", 1]], mode=False)
        changes = dict(head="new-commit", sums=b"changed", version="go1.27.2",
                       environment={"GOFLAGS": "-race", "setup-source": "script"},
                       packages=[["package", "changed-id"]], cache=[], mode=True)
        answer = key.warming_key(**inputs)
        self.assertEqual(answer, key.warming_key(**copy.deepcopy(inputs)))
        for component, value in changes.items():
            with self.subTest(component=component):
                changed = dict(inputs, **{component: value})
                self.assertNotEqual(answer, key.warming_key(**changed), component)

    def test_streamed_package_records(self):
        self.assertEqual(list(key.records(' {"ImportPath":"one"}\n{"ImportPath":"two"} ')),
                         [{"ImportPath": "one"}, {"ImportPath": "two"}])


@unittest.skipUnless(os.environ.get("ADAMIC_SETUP_INTEGRATION") == "1", "opt-in real toolchains")
class SetupIntegration(unittest.TestCase):
    def test_reruns_and_identical_uncached_answer(self):
        scratch = Path(tempfile.mkdtemp(prefix="setup-proof-"))
        print("integration logs:", scratch, flush=True)
        repository = scratch / "repository"
        (repository / "cloud").mkdir(parents=True)
        for name in ["setup.sh", "setup-key.py"]:
            shutil.copyfile(SOURCE / name, repository / "cloud" / name)
        (repository / "go.mod").write_text("module setup-proof\n\ngo 1.27\n")
        (repository / "go.sum").write_text("")
        (repository / ".gitignore").write_text("setup-proof\n")
        (repository / "main.go").write_text('package main\nimport (_ "embed"; "fmt")\n'
                                           '//go:embed greeting.txt\nvar greeting string\n'
                                           'func main() { fmt.Print(greeting) }\n')
        (repository / "greeting.txt").write_text("before\n")
        (repository / "main_test.go").write_text('package main\nimport "testing"\n'
                                                'func TestGreeting(t *testing.T) { '
                                                'if greeting == "" { t.Fatal("empty") } }\n')
        environment = os.environ.copy()
        environment["GOCACHE"] = str(scratch / "cache")
        environment["GOTOOLCHAIN"] = "auto"
        environment.pop("ADAMIC_GATE_UNCACHED", None)
        number = 0

        def run(command, expected=0, uncached=False):
            nonlocal number
            number += 1
            log = scratch / f"{number:02d}.log"
            env = dict(environment)
            if uncached:
                env["ADAMIC_GATE_UNCACHED"] = "1"
            with log.open("wb") as output:
                result = subprocess.run(command, cwd=repository, env=env, stdout=output,
                                        stderr=subprocess.STDOUT, timeout=300)
            text = log.read_text()
            if expected == 0:
                self.assertEqual(result.returncode, 0, text)
            else:
                self.assertNotEqual(result.returncode, 0, text)
            return text

        run(["git", "init"])
        run(["git", "add", "."])
        commit = ["git", "-c", "user.name=Setup proof", "-c", "user.email=setup@example.invalid",
                  "commit"]
        run(commit + ["-m", "Initial proof"])
        setup = ["bash", "cloud/setup.sh"]
        self.assertIn("go build ready", run(setup))
        warm = run(setup)
        self.assertIn("go build skipped", warm)
        self.assertIn("test binaries deferred", warm)
        (repository / "go.sum").write_text("\n")
        self.assertIn("go build ready", run(setup))
        self.assertIn("go build skipped", run(setup))
        run(commit + ["--allow-empty", "-m", "Change HEAD only"])
        self.assertIn("go build ready", run(setup))
        (repository / "greeting.txt").write_text("after\n")
        self.assertIn("go build ready", run(setup))
        (repository / "broken.go").write_text("package main\nnot Go\n")
        run(setup, expected=1)
        (repository / "broken.go").unlink()
        self.assertIn("test binaries warm", run(setup + ["--warm-tests"]))
        self.assertIn("test binaries skipped", run(setup + ["--warm-tests"]))
        cached_binary = scratch / "cached"
        uncached_binary = scratch / "uncached"
        run(["go", "build", "-o", str(cached_binary), "."])
        tools = Path(environment.get("ADAMIC_TOOLS", "/opt/adamic-tools"))
        shell_environment = (tools / "env.sh").read_bytes()
        self.assertIn("go build ready", run(setup, uncached=True))
        self.assertEqual(shell_environment, (tools / "env.sh").read_bytes())
        run(["go", "build", "-a", "-o", str(uncached_binary), "."])
        self.assertEqual(cached_binary.read_bytes(), uncached_binary.read_bytes())
        self.assertEqual(run([str(cached_binary)]), "after\n")
        self.assertEqual(run([str(uncached_binary)]), "after\n")
        # Remove an actual export artifact from this proof's isolated Go cache.
        package_log = scratch / "packages.json"
        with package_log.open("wb") as output:
            subprocess.run(["go", "list", "-export", "-json", "."], cwd=repository,
                           env=environment, stdout=output, check=True)
        artifact = Path(next(key.records(package_log.read_text()))["Export"])
        self.assertTrue(artifact.is_relative_to(scratch / "cache"))
        artifact.unlink()
        self.assertIn("go build ready", run(setup))


if __name__ == "__main__":
    unittest.main()
