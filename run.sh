set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
git -C ~/full-gate/tools fetch -q origin "ea0eba8005351e8b359f6618ee9818b598cdb281" && git -C ~/full-gate/tools switch -q --detach "ea0eba8005351e8b359f6618ee9818b598cdb281"
git -C ~/full-gate/tree fetch -q origin "5348895b4f1e4430a8ae1990f7d48dfd173bac75" && git -C ~/full-gate/tree switch -q --detach "5348895b4f1e4430a8ae1990f7d48dfd173bac75"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
taskset -c "$((cpus * 3 / 4))-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "5348895b4f1e4430a8ae1990f7d48dfd173bac75" --base "5348895b4f1e4430a8ae1990f7d48dfd173bac75" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/5348895b4f1e-20261008T044732Z"
