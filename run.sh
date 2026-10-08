set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "eaa21bb3fe9231fb16399fc0aacb55103b8a2b49" && git -C ~/full-gate/tools switch -q --detach "eaa21bb3fe9231fb16399fc0aacb55103b8a2b49"
git -C ~/full-gate/tree fetch -q origin "c6761c24c4bb35e5a7edbcbe0c5e8937d1879c3e" && git -C ~/full-gate/tree switch -q --detach "c6761c24c4bb35e5a7edbcbe0c5e8937d1879c3e"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
taskset -c "$((cpus * 3 / 4))-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "c6761c24c4bb35e5a7edbcbe0c5e8937d1879c3e" --base "c6761c24c4bb35e5a7edbcbe0c5e8937d1879c3e" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/c6761c24c4bb-20261008T032152Z"
