"""check.py's guards against judging a commit from the wrong place."""
from pathlib import Path
import subprocess
import tempfile
import unittest

import check


class CatalogDriftTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.repository = Path(self.directory.name).resolve()
        self.catalog = self.repository / "verify/catalog"
        self.catalog.mkdir(parents=True)
        self.git("init", "-q")
        self.git("config", "user.email", "catalog@test")
        self.git("config", "user.name", "catalog test")
        (self.catalog / "01-old.patch").write_text("old\n")
        (self.repository / "outside.txt").write_text("one\n")
        self.git("add", "-A")
        self.git("commit", "-q", "-m", "first")
        self.first = self.git("rev-parse", "HEAD")

    def tearDown(self):
        self.directory.cleanup()

    def git(self, *arguments):
        return subprocess.run(["git", "-C", str(self.repository), *arguments], check=True,
                              capture_output=True, text=True).stdout.strip()

    def drift(self, commit):
        return check.catalog_drift(self.repository, self.catalog, commit)

    def test_a_checkout_at_the_commit_has_no_drift(self):
        self.assertEqual(self.drift(self.first), [])

    def test_a_change_outside_the_catalog_is_not_drift(self):
        (self.repository / "outside.txt").write_text("two\n")
        self.git("commit", "-q", "-am", "outside")
        self.assertEqual(self.drift(self.first), [])

    def test_a_newer_commit_judged_from_a_stale_checkout_is_drift(self):
        (self.catalog / "01-old.patch").write_text("refreshed\n")
        self.git("commit", "-q", "-am", "refresh")
        newer = self.git("rev-parse", "HEAD")
        self.git("checkout", "-q", self.first)
        self.assertEqual(self.drift(newer), ["verify/catalog/01-old.patch"])

    def test_an_uncommitted_edit_is_drift(self):
        (self.catalog / "01-old.patch").write_text("edited\n")
        self.assertEqual(self.drift(self.first), ["verify/catalog/01-old.patch"])

    def test_an_untracked_patch_is_drift(self):
        (self.catalog / "02-new.patch").write_text("new\n")
        self.assertEqual(self.drift(self.first), ["verify/catalog/02-new.patch"])


class CgroupTests(unittest.TestCase):
    def test_a_box_without_cgroup_v2_prints_unavailable(self):
        with tempfile.TemporaryDirectory() as directory:
            self.assertEqual(check.cgroup_cpu_max(Path(directory) / "cpu.max"), "unavailable")

    def test_the_file_is_read_when_present(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "cpu.max"
            path.write_text("400000 100000\n")
            self.assertEqual(check.cgroup_cpu_max(path), "400000 100000")


if __name__ == "__main__":
    unittest.main()
