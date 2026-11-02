# Preparation commands, run in /workspace/adamic on the original box.
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
git fetch origin 33796314e5bd17ec159c0c4a92dd633552bf27ab > stage3/census/changeb-intersection-folded-meter/head-33796314/fetch.log 2>&1
git worktree add --detach /workspace/scratch/compiler-c 33796314e5bd17ec159c0c4a92dd633552bf27ab > stage3/census/changeb-intersection-folded-meter/head-33796314/worktree.log 2>&1
rmdir /workspace/scratch/compiler-c/cohere
ln -s /workspace/adamic/cohere /workspace/scratch/compiler-c/cohere
bash stage3/census/changeb-intersection-folded-meter/head-33796314/build.sh
bash stage3/census/changeb-intersection-folded-meter/head-33796314/run.sh
bash stage3/census/changeb-intersection-folded-meter/head-33796314/compare.sh
