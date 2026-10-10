// Run in an isolated checkout with the WASI backend. The requested library base
// predates that backend; this adapter preserves the exact branch fixtures and mutants.
const fs = require("node:fs");
const path = require("node:path");
const cp = require("node:child_process");
const repository = path.resolve(__dirname, "../../..");
const proof = path.resolve(process.argv[2]);
if (fs.realpathSync(proof) === fs.realpathSync(repository))
  throw new Error("use a separate proof checkout");
const log = path.resolve(process.argv[3]);
if (log.startsWith(repository + path.sep) || log.startsWith(proof + path.sep))
  throw new Error("write the proof log outside both checkouts");
const patch = cp.execFileSync(
  "git",
  ["diff", "53f44e05", "--", "internal/lower/library_math_number.go"],
  { cwd: repository },
);
cp.execFileSync("git", ["apply", "--check", "-"], { cwd: proof, input: patch });
cp.execFileSync("git", ["apply", "-"], { cwd: proof, input: patch });
for (const file of fs
  .readdirSync(path.join(repository, "internal/oracle/testdata"))
  .filter((x) => /^scout_30_tsc_.*\.a$/.test(x))) {
  fs.copyFileSync(
    path.join(repository, "internal/oracle/testdata", file),
    path.join(proof, "internal/oracle/testdata", file),
    fs.constants.COPYFILE_EXCL,
  );
}
const relative = "internal/oracle/scout_30_tsc_test.go";
let source = fs.readFileSync(path.join(repository, relative), "utf8");
source = source.replace(
  '"github.com/system-inc/adamic/internal/ir"',
  '"github.com/system-inc/adamic/internal/ir"\n"github.com/system-inc/adamic/internal/native"',
);
const marker =
  't.Log("Node alone caught wrong output; native and JavaScript exited cleanly")';
if (!source.includes(marker)) throw new Error("mutant proof marker missing");
source = source.replace(
  marker,
  'if difference := disagreement(expected,onWASI(t,native.C(program))); difference != "stdout differs" { t.Fatalf("WASI mutant: %q",difference) }\n' +
    marker,
);
fs.writeFileSync(path.join(proof, relative), source, { flag: "wx" });
const fd = fs.openSync(log, "w");
try {
  const result = cp.spawnSync(
    "go",
    [
      "test",
      "./internal/oracle",
      "-run",
      "TestWASIAgreesWithNode/internal/oracle/testdata/scout_30_tsc_|TestScout30FamilyMutants",
      "-count=1",
      "-timeout=20m",
      "-v",
    ],
    {
      cwd: proof,
      env: { ...process.env, ADAMIC_ORACLE_WASI: "1" },
      stdio: ["ignore", fd, fd],
    },
  );
  if (result.error) throw result.error;
  process.exitCode = result.status ?? 1;
} finally {
  fs.closeSync(fd);
}
