# Commands executed during preparation (logs were redirected, never piped).
cd /workspace/adamic
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
# Initial abbreviated fetch returned "couldn't find remote ref 3aec8230ab".
git fetch origin 3aec8230ab 5e5f57d3b530264cb09eba5712a8a362375aede2 > /workspace/scratch/fetch.log 2>&1
git fetch origin 5e5f57d3b530264cb09eba5712a8a362375aede2 > /workspace/scratch/fetch-b.log 2>&1
git show -s --format='%H %P %s' 5e5f57d3b530264cb09eba5712a8a362375aede2
git fetch origin 3aec8230abb53ac93b92b24f5e3ae916f9cd36a8 > /workspace/scratch/fetch-a.log 2>&1
git worktree add --detach /workspace/scratch/compiler-a 3aec8230abb53ac93b92b24f5e3ae916f9cd36a8 > /workspace/scratch/worktree-a.log 2>&1
git worktree add --detach /workspace/scratch/compiler-b 5e5f57d3b530264cb09eba5712a8a362375aede2 > /workspace/scratch/worktree-b.log 2>&1
cd /workspace/scratch/compiler-a
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /workspace/scratch/npm-ci.log 2>&1
git clone --no-checkout https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-clone.log 2>&1
git -C /workspace/scratch/upstream checkout --detach 050880ce59e30b356b686bd3144efe24f875ebc8 > /workspace/scratch/upstream-checkout.log 2>&1
# Worktrees share the exact submodules initialized by setup.
rmdir /workspace/scratch/compiler-a/cohere /workspace/scratch/compiler-b/cohere
ln -s /workspace/adamic/cohere /workspace/scratch/compiler-a/cohere
ln -s /workspace/adamic/cohere /workspace/scratch/compiler-b/cohere
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > /workspace/scratch/apply.log 2>&1
