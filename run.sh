set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "37cdaf462e6b55cc8e4c29c64672e49add136831" && git -C ~/full-gate/tools switch -q --detach "37cdaf462e6b55cc8e4c29c64672e49add136831"
git -C ~/full-gate/tree fetch -q origin "132a0ed5b4bec3dc7bc838d9dfec882e5ff9505a" && git -C ~/full-gate/tree switch -q --detach "132a0ed5b4bec3dc7bc838d9dfec882e5ff9505a"
git -C ~/full-gate/tree submodule update -q --init --recursive
chrt --idle 0 ionice -c3 nice -n 19 python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "132a0ed5b4bec3dc7bc838d9dfec882e5ff9505a" --base "132a0ed5b4bec3dc7bc838d9dfec882e5ff9505a" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/132a0ed5b4be-20261008T023132Z"
