# Do not merge this branch alone

This branch changes the closure calling convention. It must never merge alone.
It lands only through codex/closure-convention, whose current reconciliation
checkpoint is 9b6b451c3479c043501c346a24570299f069d750. Follow that branch for the
compiler-enforced convention and its acceptance evidence before integration.

Mixing this branch with unreconciled host-blockers or nested-references code can
silently put the argument count in the wrong slot. TypeScript's parser worker
observed native output 9 where Node prints 1 in native-arguments-length-value.a.
