set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "95cb3cc5e518972b3e4bf022aaa5e4a59b15b049" && git -C ~/full-gate/tools switch -q --detach "95cb3cc5e518972b3e4bf022aaa5e4a59b15b049"
git -C ~/full-gate/tree fetch -q origin "6f16a1693ff41bc102c4d9bfac83b6330277c479" && git -C ~/full-gate/tree switch -q --detach "6f16a1693ff41bc102c4d9bfac83b6330277c479"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
first=$([ "all" = quarter ] && echo $((cpus * 3 / 4)) || echo 0)
taskset -c "$first-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "6f16a1693ff41bc102c4d9bfac83b6330277c479" --base "6f16a1693ff41bc102c4d9bfac83b6330277c479" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/6f16a1693ff4-20261008T075125Z"
