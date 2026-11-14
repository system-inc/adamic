# Scout ty5myjt — checkpoint, measurements pending

No shape is recommended yet. Candidate diagnostics, full single-file census deltas,
and JavaScript byte comparisons have not been measured. This checkpoint is not a
completed scout result and must not be used to authorize an adaptation.

Base: `7534c1d7575db02bb7b6512119e0685987b7e333`, detached fresh worktree
`/workspace/scratch/scout-ty5myjt/adamic` before creating this report branch.
Shared checkout is unchanged apart from fetched Git objects/worktree registration.
Recursive shallow submodule initialization succeeded, including cohere at
`5360753e19e4417675a90e92a1823bea66b13918` and its TypeScript at
`d92d9bfee114c80be2c375d72edae966176e3a4f`; no missing credentials.

Stock TypeScript was cloned using `--depth 1 --branch v6.0.3` from Microsoft and
confirmed at `050880ce59e30b356b686bd3144efe24f875ebc8`.
`npm ci` in this worktree's stage3/api succeeded (3 packages).
`ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/scout-ty5myjt/TypeScript` is supplied
to apply.sh; adapted output and all build products are in the same scratch area.

The initial census overlay build encountered `go: creating work dir: stat
/tmp/adamic-gate: no such file or directory`; after setting GOTMPDIR, cgo also
reported `cgo: open /tmp/adamic-gate/cgo-gcc-input-1712609632: no such file or directory`.
The retry sets GOTMPDIR and TMPDIR to scratch and CGO_ENABLED=0.
No adaptation or compiler source has been edited.
