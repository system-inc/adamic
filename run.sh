set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "3f5796d00696082962fe518c62af0be09ecd2548" && git -C ~/full-gate/tools switch -q --detach "3f5796d00696082962fe518c62af0be09ecd2548"
git -C ~/full-gate/tree fetch -q origin "855d114e9b37776ec3739f25d63dbf4da968d02e" && git -C ~/full-gate/tree switch -q --detach "855d114e9b37776ec3739f25d63dbf4da968d02e"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
taskset -c "$((cpus * 3 / 4))-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "855d114e9b37776ec3739f25d63dbf4da968d02e" --base "855d114e9b37776ec3739f25d63dbf4da968d02e" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/855d114e9b37-20261008T032853Z"
