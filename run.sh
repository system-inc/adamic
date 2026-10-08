set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "7a6b1f5fd959894b8229cabe53fb978fb7536372" && git -C ~/full-gate/tools switch -q --detach "7a6b1f5fd959894b8229cabe53fb978fb7536372"
git -C ~/full-gate/tree fetch -q origin "f4efdd2369311d1420aa53fdf5c1a55bdda811d4" && git -C ~/full-gate/tree switch -q --detach "f4efdd2369311d1420aa53fdf5c1a55bdda811d4"
git -C ~/full-gate/tree submodule update -q --init --recursive
nice -n 19 python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "f4efdd2369311d1420aa53fdf5c1a55bdda811d4" --base "f4efdd2369311d1420aa53fdf5c1a55bdda811d4" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/f4efdd236931-20261008T013724Z"
