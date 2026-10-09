#!/usr/bin/env bash
# Reproduce the unpushed measurement compiler; run cloud/setup.sh first.
set -euo pipefail
repo=$(cd "$(dirname "$0")/../../../.." && pwd)
base=45487a809f89885a3fc651cd590e7dabf31362dc
runtime=04f18a2a78ab8f7576e2c92deefa5a924099dad1
scratch=${1:?usage: prepare.sh NEW_SCRATCH_DIRECTORY}
if [[ -e "$scratch" ]]; then echo "refusing existing scratch: $scratch" >&2; exit 1; fi
mkdir -p "$scratch"
scratch=$(cd "$scratch" && pwd)
checkout="$scratch/compiler"
git -C "$repo" worktree add --detach "$checkout" "$base" > "$scratch/worktree.log" 2>&1
if git -C "$checkout" merge --no-commit --no-ff "$runtime" > "$scratch/merge.log" 2>&1; then
    echo 'expected the four pinned merge conflicts' >&2; exit 1
fi
python3 - "$checkout" <<'PY'
from pathlib import Path
import subprocess, sys
root=Path(sys.argv[1])
conflicts=subprocess.check_output(['git','-C',str(root),'diff','--name-only','--diff-filter=U'],text=True).splitlines()
assert set(conflicts)=={'internal/lower/cast.go','internal/lower/expression.go','internal/oracle/counts.md','internal/oracle/oracle_test.go'},conflicts
for file in ['internal/lower/cast.go','internal/oracle/counts.md','internal/oracle/oracle_test.go']:
    (root/file).write_bytes(subprocess.check_output(['git','-C',str(root),'show','HEAD:'+file]))
f=root/'internal/lower/expression.go'; s=f.read_text(); start=s.index('<<<<<<< HEAD'); end=s.index('\n',s.index('>>>>>>>',start))
block=s[start:end]
assert block.count('enumExpression')==1 and block.count('typedArrayExpression')==1,block
s=s[:start]+'''\tif value, known, err := l.enumExpression(node); known {
		return value, err
	}
	if value, handled, err := l.typedArrayExpression(node); handled {'''+s[end:]
f.write_text(s)
subprocess.run(['git','-C',str(root),'add',*conflicts],check=True)
tree=subprocess.check_output(['git','-C',str(root),'write-tree'],text=True).strip()
print('scratch staged tree:',tree)
PY
# Empty submodule mount from worktree add; reuse setup's initialized submodules.
rmdir "$checkout/cohere"
ln -s "$repo/cohere" "$checkout/cohere"
(cd "$checkout" && go build -buildvcs=false -o "$scratch/adamic" ./cmd/adamic) > "$scratch/compiler.log" 2>&1
printf '%s\n' "$scratch/adamic"
