set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "37cdaf462e6b55cc8e4c29c64672e49add136831" && git -C ~/full-gate/tools switch -q --detach "37cdaf462e6b55cc8e4c29c64672e49add136831"
git -C ~/full-gate/tree fetch -q origin "8b3883104957bdbd712c489440fb981c3953dc49" && git -C ~/full-gate/tree switch -q --detach "8b3883104957bdbd712c489440fb981c3953dc49"
git -C ~/full-gate/tree submodule update -q --init --recursive
chrt --idle 0 ionice -c3 nice -n 19 python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "8b3883104957bdbd712c489440fb981c3953dc49" --base "8b3883104957bdbd712c489440fb981c3953dc49" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/8b3883104957-20261008T020015Z"
