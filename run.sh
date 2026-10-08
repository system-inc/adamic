set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "d01caded48041c7e963d81ffdb2257123a0c0e9c" && git -C ~/full-gate/tools switch -q --detach "d01caded48041c7e963d81ffdb2257123a0c0e9c"
git -C ~/full-gate/tree fetch -q origin "b39d8146901f3cab86f4595af07ce95db0e875e4" && git -C ~/full-gate/tree switch -q --detach "b39d8146901f3cab86f4595af07ce95db0e875e4"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
taskset -c "$((cpus * 3 / 4))-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "b39d8146901f3cab86f4595af07ce95db0e875e4" --base "b39d8146901f3cab86f4595af07ce95db0e875e4" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/b39d8146901f-20261008T060744Z"
