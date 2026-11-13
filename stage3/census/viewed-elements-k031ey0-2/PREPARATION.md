# Preparation and exact measurement commands

The existing worktrees from the first measurement were reused: `/workspace/k031ey0/main` has compiler sources from c07d3b5204082b9d862091c9e53ce241ab65670d (plus the first evidence commit, which affects only its evidence directory), and `/workspace/k031ey0/head` is a16bedb211dab5584147af0903b8c6ae18fd8fde. Both use cohere 5360753e19e4417675a90e92a1823bea66b13918 and cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f. The same adapted tree `/workspace/k031ey0/adapted` from the first run was reused unchanged. It was built by main's stage3/apply.sh from upstream TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. The original adaptation commands and logs are in branch typescript/viewed-elements-log-k031ey0, commit 06b9fe01617af0f43a4a48d86fe755595a516c08.

Production internal/lower/lower.go constructs `key := &viewElementKey{}` before lowerWithElementKey; no additional fields are initialized. The seam uses exactly that empty construction on the registration lowering. No other measurement code was changed.

The seam was applied with this Python operation for each side:

```python
from pathlib import Path
for side in ('main', 'head'):
    p = Path('/workspace/k031ey0') / side / 'stage3/census/latent/units.go.txt'
    text = p.read_text()
    needle = '\t\tbase := fresh()\n'
    assert text.count(needle) == 1
    p.write_text(text.replace(needle, needle + '\t\tbase.elementKey = &viewElementKey{}\n'))
```

Then `git diff -- stage3/census/latent/units.go.txt` was saved from each worktree as main-seam.diff and head-seam.diff. Their bytes were asserted equal; seam.diff is their common exact diff. It can also be applied to either original tree with `git apply seam.diff`.

The exact sequential build and run script is measure.sh, executed as:

```bash
cd /workspace/k031ey0/main
bash /workspace/k031ey0-2/evidence/measure.sh
```

It regenerates both latent and entry overlays from each side's own sources and rebuilds separate binaries (Go go1.27.1 linux/amd64, -buildvcs=false, -p=1). It runs the same adapted tsc entry and exact single-file selection as before, with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, ADAMIC_VIEWED_ELEMENTS_LOG=1. Full mode is the mode honoring LATENT_ONLY_FILE. No workloads run concurrently. Each census has a 900-second timeout with --preserve-status, and its own exit code is printed and saved. Stderr is saved directly and never piped through head or tail.

Line counts are produced by `python3 count.py > counts.json`; all complete stderr lines beginning viewed-elements: are parsed and counted.
