# Queue box amendment proof checkpoint

Both full applies pass on frozen main f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.
The complete 746-file source comparison changes only createQueue in core.ts.
The whole lane and all 398 physically instrumented units have identical operation counts and zero differences.
The pristine contract and migration contract pass; replay preserves bytes and mtime; the owner has zero casts.
The measured and applied JavaScript tokens are identical. isEmpty and truncation are unchanged.
Baseline adamic types passes all 79 roots; baseline full core.ts census passes.

The fresh after compiler build, paired replay, after types/census/mutant and uninstrumented bounded lane are pending.
No pending result is claimed as passed. The original 99s proofs remain historical.
Raw results are packed in queue-box-evidence.tar.gz, covered by internal SHA256SUMS and the adjacent archive hash.
