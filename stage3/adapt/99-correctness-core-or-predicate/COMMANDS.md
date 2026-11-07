# Baseline census reproduction

These commands use /workspace/adamic as the task checkout. Setup and dependency
installs completed before measurement. The source was fetched shallowly at the
exact pin rather than cloning upstream history. The baseline apply ran from a
separate main worktree and had no predicate adaptation directory.

```sh
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /workspace/scratch/or-predicate-setup-resumed.log 2>&1
# Capture and report that child's exit status before proceeding.
npm ci --prefix stage3/api
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
# /workspace/typescript-source HEAD must be exactly:
# 050880ce59e30b356b686bd3144efe24f875ebc8

# /workspace/or-predicate-main is the detached origin/main worktree recorded
# in evidence/provenance.json. Its API node_modules points to the task's
# validated stage3/api/node_modules install.
bash /workspace/or-predicate-main/stage3/apply.sh --build-product /workspace/or-predicate-before \
  > /workspace/scratch/or-predicate-apply-before.log 2>&1
# Actual wrapper reported: main apply exit 0.

python3 stage3/census/latent/make_overlay.py \
  /workspace/adamic /workspace/scratch/or-predicate/overlay \
  > /workspace/scratch/or-predicate/overlay.log 2>&1
```

Every generated .go output was renamed to .go.txt, with overlay.json's
replacement paths updated. The exact resulting overlay, including the generated
state copier and refusal instrumentation, is archived. No production compiler
file was edited.

```sh
go build -buildvcs=false \
  -overlay=/workspace/scratch/or-predicate/overlay/overlay.json \
  -o /workspace/scratch/or-predicate/latent ./stage3/census/latent/tool \
  > /workspace/scratch/or-predicate/build-census.log 2>&1
# Actual wrapper reported: census build exit 0.

# run.py was executed from scratch; the shipped copy has the same bytes.
# It records the child's own exit code and stops after 30 minutes with no output.
LATENT_FULL=1 \
LATENT_ONLY_FILE=/workspace/or-predicate-before/src/compiler/core.ts \
LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_TIMES=/workspace/scratch/or-predicate/unit-times-before.jsonl \
python3 /workspace/scratch/or-predicate/run.py census-before \
  /workspace/scratch/or-predicate/latent \
  /workspace/or-predicate-before/src/compiler \
  /workspace/scratch/or-predicate/census-before.jsonl
# Actual child exit: 0. No other LATENT_ variables were set.
```

The separate stock-checker inventory uses the upstream package-lock install in
/workspace/or-predicate-packages, including its Node, Mocha and Chai type
packages. It resolves symbols and signatures, rather than matching calls by
textual callee name alone. It asserts the exact ordered arguments, occurrences,
overload assignments and the 19 predicate sites against the reference commit.

```sh
python3 /workspace/scratch/or-predicate/run.py calls-before \
  node /workspace/scratch/or-predicate/calls.cjs \
  /workspace/or-predicate-before \
  /workspace/scratch/or-predicate/calls-before.json before
# Actual child exit: 0. Program diagnostic count: 25, retained verbatim.

python3 /workspace/scratch/or-predicate/analyze.py \
  /workspace/scratch/or-predicate/census-before.jsonl \
  /workspace/scratch/or-predicate/or-declarations.json \
  /workspace/scratch/or-predicate/census-before-summary.json
```

or-declarations.json was produced by stock TypeScript 6.0.3's createSourceFile,
selecting top-level FunctionDeclaration nodes named or and recording getStart,
end, line, endLine, body presence and getText. It is archived with the raw
census. The summary tests both unit attribution and source spans, and separately
searches the entire core finding stream for an overload 1 of or reason.

The required gate is false, so no adapter, candidate tree, after census or mutant
was built. All evidence was packed only after the commands finished. Every
archive member was read back and verified against the internal SHA256SUMS.
