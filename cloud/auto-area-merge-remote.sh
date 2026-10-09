#!/usr/bin/env bash
# Sent over SSH to Cloud as ahra. Integration owns the merge implementation.
set -euo pipefail
area=$1 branch=$2 sha=$3 integration=$4 relativeOut=$5
shift 5
[ "$(id -un)" = ahra ] || { echo 'refused: area merges require Cloud user ahra'; exit 2; }
source ~/adamic-tools/env.sh
export PATH="${HOME}/adamic-tools/bin:${PATH}"
export PATH="${ADAMIC_TYPESCRIPT_SOURCE}/bin:${PATH}"
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
mkdir -p ~/area-merge
for directory in integration tree; do
  if [ ! -d ~/area-merge/${directory}/.git ]; then
    git clone -q https://github.com/system-inc/adamic.git ~/area-merge/${directory}
  fi
done
# Pin the trusted tools to the same integration commit used for routing on the Mac.
git -C ~/area-merge/integration fetch -q origin "+refs/heads/cloud/merge-tree:refs/remotes/origin/cloud/merge-tree" "${integration}"
[ -z "$(git -C ~/area-merge/integration status --porcelain)" ] || { echo 'refused: remote integration checkout has local changes'; exit 1; }
git -C ~/area-merge/integration switch -q --detach "${integration}"
export ADAMIC_AREA_WORKTREES=${HOME}/area-merge/worktrees
export PYTHONDONTWRITEBYTECODE=1
export GIT_AUTHOR_NAME=kirkouimet GIT_AUTHOR_EMAIL=kirk@kirkouimet.com
export GIT_COMMITTER_NAME=kirkouimet GIT_COMMITTER_EMAIL=kirk@kirkouimet.com
out=${HOME}/${relativeOut}
mkdir -p "${out}"
cd ~/area-merge/tree
set +e
taskset -c 0-23 bash ~/area-merge/integration/cloud/integration/area-merge.sh "${area}" "${branch}" "${sha}" "$@" > "${out}/output.log" 2>&1
code=$?
set -e
cat "${out}/output.log"
# Bundle every transaction's printed log directory and individually referenced logs.
# The manifest lets the Mac rewrite remote paths to copied paths for worker feedback.
python3 - "${out}" <<'LOGS' || echo 'area-merge: could not bundle all remote logs'
import json, pathlib, re, shutil, sys, tempfile
out = pathlib.Path(sys.argv[1])
text = (out / 'output.log').read_text(errors='replace')
paths = set(re.findall(r'logs in (/[^\s;,)]+)', text))
paths.update(str(pathlib.Path(name).parent) for name in re.findall(r'(/[^\s();,]+\.(?:log|txt))', text))
# Only mktemp transaction directories under this process's temp root are log bundles;
# a path mentioned in a test diagnostic must not copy an arbitrary filesystem tree.
temporary = pathlib.Path(tempfile.gettempdir()).resolve()
roots = set()
for name in paths:
    try:
        relative = pathlib.Path(name).resolve().relative_to(temporary)
    except ValueError:
        continue
    if relative.parts and relative.parts[0].startswith('tmp.'):
        roots.add(str(temporary / relative.parts[0]))
mapping = {}
for index, name in enumerate(sorted(roots)):
    source = pathlib.Path(name)
    if source.is_dir():
        relative = 'logs/' + str(index)
        shutil.copytree(source, out / relative)
        mapping[name] = relative
(out / 'logs.json').write_text(json.dumps(mapping))
LOGS
exit "${code}"
