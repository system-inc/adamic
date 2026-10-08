set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
git -C ~/full-gate/tools fetch -q origin "4af2c103d4e5e0f5de3a969fed5d4a3a7ea4276a" && git -C ~/full-gate/tools switch -q --detach "4af2c103d4e5e0f5de3a969fed5d4a3a7ea4276a"
git -C ~/full-gate/tree fetch -q origin "749a69adbfae2a7bf22c1f0436d9ef73345c3070" && git -C ~/full-gate/tree switch -q --detach "749a69adbfae2a7bf22c1f0436d9ef73345c3070"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
first=$([ "all" = quarter ] && echo $((cpus * 3 / 4)) || echo 0)
taskset -c "$first-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "749a69adbfae2a7bf22c1f0436d9ef73345c3070" --base "749a69adbfae2a7bf22c1f0436d9ef73345c3070" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/749a69adbfae-20261008T100838Z"
