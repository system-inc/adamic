import os, time, subprocess, json
from pathlib import Path
class Harness:
    def aCheck(self, paths):
        """Each changed .a outside a package through stage 0's front end (adamic c: the checker's proven
        types, then the refusal pass; it stops before C): a clean result or a "can't lower ... yet" stop
        passes, since nothing was refused. A file whose first line is "// a-check: refused <rule>" or
        "// a-check: type error <code>" must fail exactly that way instead."""
        started = time.monotonic()
        adamic = os.path.realpath(self.arguments.tree) + "-binaries/adamic-acheck"
        if self.step("a-check-build", ["go", "build", "-o", adamic, "./cmd/adamic"]) is False:
            self.exits["a-check"] = 1
            return
        results, failed = {}, []
        for path in paths:
            with open(os.path.join(self.arguments.tree, path), errors="replace") as handle:
                header = handle.readline().strip()
            expect = header[len("// a-check:"):].strip() if header.startswith("// a-check:") else ""
            process = self.spawn([adamic, "c", path], subprocess.PIPE, subprocess.PIPE)
            _, errors = process.communicate()
            if process.returncode == 0 or ("can't lower" in errors and " yet" in errors):
                outcome = "checked"
            elif "Adamic 0.1 refuses" in errors:
                outcome = "refused"
            elif " error TS" in errors:
                outcome = "type error"
            else:
                outcome = "failed"
            first = errors.strip().splitlines()[0] if errors.strip() else ""
            if expect.startswith("refused"):
                ok = outcome == "refused" and expect[len("refused"):].strip() in errors
            elif expect.startswith("type error"):
                code = expect[len("type error"):].strip()
                ok = outcome == "type error" and (not code or ("error " + code) in errors)
            else:
                ok = outcome == "checked"
            results[path] = {"outcome": outcome, "expected": expect or "checked", "first": first[:300]}
            if not ok:
                failed.append("%s: %s, expected %s (%s)" % (path, outcome, expect or "checked", first[:200]))
        self.result["a_check"] = results
        self.steps["a-check"] = round(time.monotonic() - started, 1)
        self.exits["a-check"] = 1 if failed else 0
        if failed:
            self.fail("a-check", "a-check failed for %d file%s:\n%s" % (len(failed), "" if len(failed) == 1 else "s", "\n".join(failed[:50])))

    def __init__(self):
        self.arguments=type('Args',(),{'tree':os.getcwd()})();self.result={};self.steps={};self.exits={};self.failure=None
    def step(self,name,cmd):
        with open(os.environ.get('ADAMIC_ACHECK_BUILD_LOG', '/tmp/adamic-acheck-build.log'), 'w') as log:
            return subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT).returncode==0
    def spawn(self,cmd,stdout,stderr):
        return subprocess.Popen(cmd,stdout=stdout,stderr=stderr,text=True)
    def fail(self,name,detail): self.failure=detail
path='stage3/verdict/fixtures/case-host/input.a'
raw=Path(path).read_bytes()
h=Harness();h.aCheck([path]);print('original',h.exits,h.result);assert h.exits['a-check']==0
try:
    Path(path).write_bytes(raw.replace(b'type error TS2322',b'type error TS9999',1))
    m=Harness();m.aCheck([path]);print('mutant',m.exits,m.failure);assert m.exits['a-check']==1
finally: Path(path).write_bytes(raw)
r=Harness();r.aCheck([path]);print('restored',r.exits);assert r.exits['a-check']==0
