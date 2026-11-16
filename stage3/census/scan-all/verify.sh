#!/usr/bin/env bash
set -euo pipefail
repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)  # no git: a sealed job's tree has no .git
cd "$repo"
: "${SCAN_ALL_SCRATCH:?set a scratch directory outside the checkout}"
mkdir -p "$SCAN_ALL_SCRATCH"
python3 stage3/census/scan-all/make_overlay.py "$SCAN_ALL_SCRATCH/overlay"
go build -overlay "$SCAN_ALL_SCRATCH/overlay/overlay.json" -o "$SCAN_ALL_SCRATCH/scan-all" ./stage3/census/scan-all/tool/main.go
python3 - "$repo" "$SCAN_ALL_SCRATCH" <<'PYTHON'
import json,pathlib,sys
repo,scratch=map(pathlib.Path,sys.argv[1:])
(scratch/'production-overlay.json').write_text(json.dumps({'Replace':{str(repo/'stage3/census/scan-all/production/main.go'):str(repo/'stage3/census/scan-all/production.go.txt')}}))
PYTHON
go build -overlay "$SCAN_ALL_SCRATCH/production-overlay.json" -o "$SCAN_ALL_SCRATCH/production" ./stage3/census/scan-all/production/main.go
export SCAN_ALL_BINARY="$SCAN_ALL_SCRATCH/scan-all" SCAN_ALL_PRODUCTION="$SCAN_ALL_SCRATCH/production"
go test -v -count=1 -timeout=60s ./stage3/census/scan-all
controls=(count-stop count-clean rest-stop rest-clean record-stop record-clean cycle-stop cycle-clean)
for name in "${controls[@]}"; do
  (cd stage3/census/scan-all && timeout 55 python3 controls.py "$name") > "$SCAN_ALL_SCRATCH/control-$name.log" 2>&1 \
    || { cat "$SCAN_ALL_SCRATCH/control-$name.log" >&2; echo "control failed: $name" >&2; exit 1; }
  echo "control passed: $name"
done
# Each first-error mutant restores production's refusal scan; its control must fail
# because the later target disappears or the clean traversal no longer completes.
for name in "${controls[@]}"; do
  if (cd stage3/census/scan-all && SCAN_ALL_MUTANT_FIRST_ERROR=1 timeout 55 python3 controls.py "$name") > "$SCAN_ALL_SCRATCH/mutant-$name.log" 2>&1; then
    echo "first-error mutant survived: $name" >&2; exit 1
  fi
  grep -Eq "continuation-target|continuation-completion" "$SCAN_ALL_SCRATCH/mutant-$name.log" \
    || { cat "$SCAN_ALL_SCRATCH/mutant-$name.log" >&2; echo "mutant failed for another reason: $name" >&2; exit 1; }
  echo "first-error mutant red: $name"
done

# These operator proofs run the unchanged named Go test on isolated source copies.
# The test itself only runs Python --check, never Go or a compiler build.
python3 - "$repo" "$SCAN_ALL_SCRATCH" <<'PYTHON'
import os, pathlib, shutil, subprocess, sys, tempfile
repo, scratch = map(pathlib.Path, sys.argv[1:])
name = 'TestScanAllMainAnchorsMatchExactly'
lower_anchor = 'lowering.refuse(module)'
refusal_anchor = 'contracts = func(node *ast.Node) bool {'
signature = 'func (l *lowering) refuse(module *ast.SourceFile) error {'
files = ['internal/lower/lower.go', 'internal/lower/refusals.go',
         'stage3/census/scan-all/make_overlay.py', 'stage3/census/scan-all/scan_all_test.go']
real_go = shutil.which('go')
with tempfile.TemporaryDirectory(prefix='scan-all-anchor-copies-', dir=scratch) as temporary:
    temporary = pathlib.Path(temporary)
    trap = temporary/'no-go'; trap.mkdir()
    marker = temporary/'go-was-run'
    fake_go = trap/'go'
    fake_go.write_text('#!/bin/sh\nprintf "unexpected go subprocess\\n" > "$SCAN_ALL_GO_MARKER"\nexit 97\n')
    fake_go.chmod(0o755)
    env = dict(os.environ, GOWORK='off', SCAN_ALL_GO_MARKER=str(marker),
               PATH=str(trap)+os.pathsep+os.environ['PATH'])
    cases = [('lower-removed', lower_anchor), ('lower-duplicated', lower_anchor),
             ('refusal-removed', refusal_anchor), ('refusal-duplicated', refusal_anchor),
             ('pipeline-empty', 'pipeline-site'), ('unrelated-comment', None)]
    for case, expected in cases:
        root = temporary/case
        for filename in files:
            target = root/filename
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(repo/filename, target)
        (root/'go.mod').write_text('module scanall-anchor-proof\n\ngo 1.27\n')
        lower = root/'internal/lower/lower.go'
        refusal = root/'internal/lower/refusals.go'
        if case == 'lower-removed':
            lower.write_text(lower.read_text().replace(lower_anchor, 'lowering.removedAnchor(module)', 1))
        elif case == 'lower-duplicated':
            lower.write_text(lower.read_text()+'\n// duplicate anchor: '+lower_anchor+'\n')
        elif case == 'refusal-removed':
            refusal.write_text(refusal.read_text().replace(refusal_anchor, 'contractsMissing = func(node *ast.Node) bool {', 1))
        elif case == 'refusal-duplicated':
            refusal.write_text(refusal.read_text().replace(signature, signature+'\n// duplicate anchor: '+refusal_anchor, 1))
        elif case == 'pipeline-empty':
            lower.write_text(lower.read_text().replace('if err :=', 'if missingErr :='))
        else:
            lower.write_text('// Unrelated comment added; no anchor changed.\n'+lower.read_text())
        run = subprocess.run([real_go, 'test', '-v', '-count=1', '-timeout=60s',
                              './stage3/census/scan-all', '-run', '^'+name+'$'],
                             cwd=root, env=env, text=True, capture_output=True, timeout=60)
        output = run.stdout+run.stderr
        (scratch/('anchor-'+case+'.log')).write_text(output)
        assert not marker.exists(), (case, 'gate check launched Go', output)
        if expected:
            assert run.returncode != 0 and '--- FAIL: '+name in output and expected in output, (case, output)
            print(case+': '+name+' RED; names '+expected)
        else:
            assert run.returncode == 0 and '--- PASS: '+name in output, (case, output)
            print(case+': '+name+' GREEN')
    print('anchor checks launched no Go subprocess; all source mutations stayed in scratch')
PYTHON
