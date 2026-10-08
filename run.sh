set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
git -C ~/full-gate/tools fetch -q origin "0fb8e96feec631126538f1342c9648d06e2bdade" && git -C ~/full-gate/tools switch -q --detach "0fb8e96feec631126538f1342c9648d06e2bdade"
git -C ~/full-gate/tree fetch -q origin "ef3141e9b1152ab51b51497f8ce3a2799449c8a3" && git -C ~/full-gate/tree switch -q --detach "ef3141e9b1152ab51b51497f8ce3a2799449c8a3"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
first=$([ "all" = quarter ] && echo $((cpus * 3 / 4)) || echo 0)
taskset -c "$first-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "ef3141e9b1152ab51b51497f8ce3a2799449c8a3" --base "ef3141e9b1152ab51b51497f8ce3a2799449c8a3" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/ef3141e9b115-20261008T093452Z"
