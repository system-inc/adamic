#!/usr/bin/env python3
"""Exercise warming keys and real setup invalidation; all subprocess output goes to logs."""
import copy
import contextlib
import io
import json
import importlib.util
import os
from pathlib import Path
import shutil
import sys
import subprocess
import tempfile
import unittest
from unittest.mock import patch

sys.dont_write_bytecode = True
SOURCE = Path(__file__).resolve().parent
KEY_MODULE = Path(os.environ.get("ADAMIC_SETUP_KEY_MODULE", SOURCE / "setup-key.py"))
spec = importlib.util.spec_from_file_location("setup_key", KEY_MODULE)
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

    def test_module_and_workspace_manifests(self):
        with tempfile.TemporaryDirectory(prefix="setup-manifests-") as directory:
            module = Path(directory) / "go.mod"
            sums = Path(directory) / "go.sum"
            workspace = Path(directory) / "go.work"
            workspace_sums = Path(directory) / "go.work.sum"
            paths = {str(module), str(sums), str(workspace), str(workspace_sums)}
            module.write_text("module proof\n")
            previous = key.manifest_bytes(paths)
            for file in [sums, workspace, workspace_sums]:
                file.write_text("")
                answer = key.manifest_bytes(paths)
                self.assertNotEqual(previous, answer, "missing and empty must differ")
                previous = answer
            for file in [module, sums, workspace, workspace_sums]:
                file.write_text(file.read_text() + "\n")
                answer = key.manifest_bytes(paths)
                self.assertNotEqual(previous, answer, str(file))
                previous = answer

    def test_collected_manifests_invalidate(self):
        with tempfile.TemporaryDirectory(prefix="setup-collected-") as directory:
            repository = Path(directory)
            (repository / "cloud").mkdir()
            (repository / "cloud/setup.sh").write_text("setup source")
            cache = repository / "cache"
            cache.mkdir()
            artifact = cache / "artifact-d"
            artifact.write_text("compiled dependency")
            dependency = repository / "dependency"
            dependency.mkdir()
            paths = [repository / "go.mod", repository / "go.sum", repository / "go.work",
                     repository / "go.work.sum", dependency / "go.mod", dependency / "go.sum",
                     repository / "cloud/markdown-width/package.json",
                     repository / "cloud/markdown-width/package-lock.json",
                     repository / "cloud/markdown-width/npm-bootstrap.json",
                     repository / "cloud/setup-markdown-width.py", repository / "cloud/setup-stage3-api.py",
                     repository / "stage3/api/package.json", repository / "stage3/api/package-lock.json"]
            for file in paths:
                file.parent.mkdir(parents=True, exist_ok=True)
                file.write_text("manifest\n")
            packages = repository / "packages.json"
            packages.write_text(json.dumps({"ImportPath": "proof", "BuildID": "unchanged",
                                           "Export": str(artifact),
                                           "Module": {"GoMod": str(dependency / "go.mod")}}))
            environment = {"GOGCCFLAGS": "-fPIC", "GOCACHE": str(cache),
                           "GOWORK": str(repository / "go.work")}

            def command(arguments):
                if arguments == ["go", "env", "-json"]:
                    return json.dumps(environment).encode()
                if arguments == ["go", "version"]:
                    return b"go version go1.27.1 linux/amd64\n"
                if arguments == ["git", "rev-parse", "HEAD"]:
                    return b"unchanged-head\n"
                raise AssertionError(arguments)

            def answer():
                original = Path.cwd()
                try:
                    with patch.object(key.subprocess, "check_output", command), patch.object(
                        key.sys, "argv", ["setup-key.py", str(repository), str(packages), "false"]
                    ), contextlib.redirect_stdout(io.StringIO()) as output:
                        key.main()
                    return output.getvalue()
                finally:
                    os.chdir(original)

            previous = answer()
            for file in paths:
                file.write_text(file.read_text() + "changed\n")
                current = answer()
                self.assertNotEqual(previous, current, str(file))
                previous = current

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
        for name in ["setup.sh", "setup-key.py", "setup-markdown-width.py", "setup-stage3-api.py", "setup-modules.py"]:
            shutil.copyfile(KEY_MODULE if name == "setup-key.py" else SOURCE / name,
                            repository / "cloud" / name)
        shutil.copytree(SOURCE / "markdown-width", repository / "cloud/markdown-width")
        (repository / "internal/boundedrun").mkdir(parents=True)
        for name in ["shell.sh", "python.py"]:
            shutil.copyfile(SOURCE.parent / "internal/boundedrun" / name,
                            repository / "internal/boundedrun" / name)
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
        # The stamp's HEAD and checksum checks must stand on their own, rather than being
        # incidentally caught by VCS metadata in main-package compilation actions.
        environment["GOFLAGS"] = "-buildvcs=false"
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
