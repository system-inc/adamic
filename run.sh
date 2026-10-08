set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
# Stock tsc is the gates' pinned TypeScript 6.0.3 (the LKG checkout setup provides), for oracles that
# take it from PATH (cmd/adamic-test262's stock-rejection check). No box had a tsc on PATH.
export PATH="${ADAMIC_TYPESCRIPT_SOURCE}/bin:${PATH}"
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
git -C ~/full-gate/tools fetch -q origin "38ef90ede4ddc25b5d3eaad83bdd0b2eec7b2ee2" && git -C ~/full-gate/tools switch -q --detach "38ef90ede4ddc25b5d3eaad83bdd0b2eec7b2ee2"
# A declared tool the box lacks (cloud/fast-gate/tools.txt) makes the run void, naming the box, never red.
if ! lacks=$(bash ~/full-gate/tools/cloud/fast-gate/tools-check.sh); then
  echo "${lacks}" > ~/"full-gate/out/ff36b67c4066-20261008T123457Z"/tools-missing.txt
  echo "void: ff36b67c406678b7d0e1ef5e08187664ce97d826 full gate, box $(hostname) lacks a declared tool: $(echo "${lacks}" | sed -E 's/^lacks ([^:]+):.*/\1/' | tr '\n' ' ')" > ~/"full-gate/out/ff36b67c4066-20261008T123457Z"/status.txt
  echo '{"finished": true, "void": true}' > ~/"full-gate/out/ff36b67c4066-20261008T123457Z"/full.json
  exit 0
fi
git -C ~/full-gate/tree fetch -q origin "ff36b67c406678b7d0e1ef5e08187664ce97d826" && git -C ~/full-gate/tree switch -q --detach "ff36b67c406678b7d0e1ef5e08187664ce97d826"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
first=$([ "all" = quarter ] && echo $((cpus * 3 / 4)) || echo 0)
taskset -c "$first-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "ff36b67c406678b7d0e1ef5e08187664ce97d826" --base "ff36b67c406678b7d0e1ef5e08187664ce97d826" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/ff36b67c4066-20261008T123457Z"
