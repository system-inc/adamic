# Setup-fast merge and TypeScript Prettier input

Merged origin/area/developer-tools at 5794c87661b030d24cee457bbec4900f260e025d into devtools/setup-fast at 146e72c93b49581e8f8f63bec8233284392d8e5b. Merge commit 8c31c1d has both as parents. No rebasing. The conflicts preserve stage3 API npm preparation, gate npm preparation, Go readiness and module downloads, and both helpers in the setup test fixture. Stage3 preparation now also uses the area's bounded subprocess runner. The markdown width harness fix is retained.

ADAMIC_TS_PRETTIER shares the css-printer directory and its checked-in exact prettier 3.9.6 lockfile/SHA512 integrity. Both it and ADAMIC_CSS_PRINTER_LIBRARY export the same npm install root under --gate-inputs. npm directory enumeration is deduplicated so it prepares the seat only once. Ordinary setup unsets the new variable. No new cache or lockfile is introduced.

Read the tsprinter doc/expressions loaders and README from origin/stage1-format/ts-printer: they require createRequire(root + '/package.json') and assert Prettier 3.9.6. Their actual loaders passed a document and a TypeScript expression with the existing shared install and zero stderr. This checks placement/version; the full TypeScript printer suite is not present on this developer-tools branch and was not run.

ADAMIC_TYPESCRIPT_SOURCE already exports the typescript directory. Installer TS_COMMIT and the actual installed checkout HEAD both equal 050880ce59e30b356b686bd3144efe24f875ebc8.

Commands and results:

- python3 cloud/test_setup.py: 4 unit checks passed; real integration opt-in.
- python3 cloud/test_markdown_setup.py: 4 unit checks passed; two opt-in.
- python3 cloud/test_stage3_setup.py: 3 unit checks passed; one opt-in.
- python3 cloud/test_setup_modules.py: 5 passed.
- python3 cloud/test_gate_inputs.py: 6 unit checks passed; archive integration opt-in.
- ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_setup.py with ADAMIC_TOOLS at an isolated /tmp/adamic-gate/merge-fast-tools directory reusing the installed Go/LLVM/Node: all 5 passed, including cached/uncached identical binary output and real invalidation. The first attempt using shared workspace tools hit disk exhaustion; the isolated retry passed.
- ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_stage3_setup.py: all 4 passed.
- ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_markdown_setup.py: all 6 passed.
- ADAMIC_SETUP_INTEGRATION=1 ADAMIC_MARKDOWN_SETUP_MODULE=/workspace/adamic/cloud/setup-gate-npm.py python3 cloud/test_markdown_setup.py: all 6 passed, including checksum and installed-byte validation, real lock changes and uncached equality.
- Missing ADAMIC_TS_PRETTIER export mutant: caught by test_shared_prettier_exports_and_typescript_pin.
- Duplicate css-printer install mutant (remove directory deduplication): caught by test_shared_prettier_installed_once.
- bash -n cloud/setup.sh and git diff --check: passed.

Test output is retained in adjacent log files. Full gate and large formatter corpora were not rerun. No main push or pull request.
