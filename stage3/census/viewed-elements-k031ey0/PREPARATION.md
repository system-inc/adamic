# Exact preparation commands

Executed sequentially. Initial repository: `/workspace/adamic`. No AGENTS.md files were found in the workspace (excluding dependencies).

```bash
cd /workspace/adamic
git fetch origin scratch/lower/viewed-elements-log-main scratch/lower/viewed-elements-log-head
mkdir -p /workspace/k031ey0
git worktree add --detach /workspace/k031ey0/main c07d3b5204
git worktree add --detach /workspace/k031ey0/head a16bedb211
git diff --stat c07d3b5204 a16bedb211
git -C /workspace/k031ey0/main submodule update --init --recursive --depth 1
cd /workspace/k031ey0/head
git -c protocol.file.allow=always -c url.file:///workspace/k031ey0/main/cohere.insteadOf=https://github.com/system-inc/cohere.git -c url.file:///workspace/k031ey0/main/cohere/TypeScript.insteadOf=https://github.com/system-inc/TypeScript.git submodule update --init --recursive --depth 1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/k031ey0/upstream
cd /workspace/k031ey0/main/stage3/api
mkdir -p /workspace/k031ey0/evidence
npm ci --ignore-scripts --no-audit --no-fund > /workspace/k031ey0/evidence/packages.log 2>&1
source /workspace/adamic-tools/env.sh
mkdir -p "$TMPDIR"
ADAMIC_TYPESCRIPT_SOURCE=/workspace/k031ey0/upstream bash /workspace/k031ey0/main/stage3/apply.sh /workspace/k031ey0/adapted > /workspace/k031ey0/evidence/apply.stdout 2> /workspace/k031ey0/evidence/apply.stderr
cd /workspace/k031ey0/main
git diff --exit-code c07d3b5204 a16bedb211 -- stage3
bash /workspace/k031ey0/measure.sh
```

`measure.sh` is included verbatim. It builds only the entry product using the latent overlay followed by the meter entry overlay, as in build-probes.py/twice-daily.sh. `-p=1` serializes Go package compilation. `timeout --preserve-status 900` bounds each census and retains its child status (143 on timeout termination); each numeric status is also printed and saved. No stderr stream is piped or truncated. Full mode is required for LATENT_ONLY_FILE, per latent/README.md, File selection and program record facts.

Pins: main c07d3b5204082b9d862091c9e53ce241ab65670d; head a16bedb211dab5584147af0903b8c6ae18fd8fde; cohere 5360753e19e4417675a90e92a1823bea66b13918; cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f; upstream 050880ce59e30b356b686bd3144efe24f875ebc8. Go go1.27.1 linux/amd64.
