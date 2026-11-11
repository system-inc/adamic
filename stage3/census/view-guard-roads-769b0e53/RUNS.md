# Child exits, elapsed time and peak memory

Elapsed values are monotonic seconds. Peak RSS is wait4 ru_maxrss in KiB. Census binaries are single processes; preparation/batch wrappers report the wait4 high-water value, not simultaneous descendant memory summed together. Runs overlap, so these are not isolated compiler performance comparisons. first-parent is the historical intentionally stopped run from the preceding blocked report and is excluded from this census.

| Run | Child exit | Seconds | Peak RSS KiB | LATENT_ONLY_FILE |
| --- | ---: | ---: | ---: | --- |
| [apply](raw/apply.json) | 0 | 1310.184 | 2473852 |  |
| [baseline-head](raw/baseline-head.json) | 0 | 4869.733 | 5123252 |  |
| [baseline-parent](raw/baseline-parent.json) | 0 | 4966.145 | 5101784 |  |
| [build-guard-head](raw/build-guard-head.json) | 0 | 115.769 | 1455932 |  |
| [build-guard-parent](raw/build-guard-parent.json) | 0 | 115.672 | 1455556 |  |
| [build-head](raw/build-head.json) | 0 | 675.184 | 2283552 |  |
| [build-parent](raw/build-parent.json) | 0 | 678.967 | 2364420 |  |
| [control-head-base](raw/control-head-base.json) | 0 | 0.133 | 31240 |  |
| [control-head-mutant](raw/control-head-mutant.json) | 0 | 0.170 | 32080 |  |
| [control-head-real](raw/control-head-real.json) | 0 | 0.133 | 32144 |  |
| [control-parent-base](raw/control-parent-base.json) | 0 | 0.134 | 31684 |  |
| [control-parent-mutant](raw/control-parent-mutant.json) | 0 | 0.135 | 32196 |  |
| [control-parent-real](raw/control-parent-real.json) | 0 | 0.141 | 31848 |  |
| [control-parent](raw/control-parent.json) | 0 | 0.143 | 31944 |  |
| [first-parent](raw/first-parent.json) | -15 | 571.018 | 1330972 |  |
| [full-01e6ae4837-parent-base](raw/full-01e6ae4837-parent-base.json) | 0 | 146.783 | 1107384 | src/compiler/transformers/esnext.ts |
| [full-01e6ae4837-parent-guard](raw/full-01e6ae4837-parent-guard.json) | 0 | 143.572 | 1136568 | src/compiler/transformers/esnext.ts |
| [full-0854c1cf41-head-guard](raw/full-0854c1cf41-head-guard.json) | 0 | 74.474 | 1000052 | src/compiler/symbolWalker.ts |
| [full-0854c1cf41-parent-base](raw/full-0854c1cf41-parent-base.json) | 0 | 68.144 | 1000272 | src/compiler/symbolWalker.ts |
| [full-0854c1cf41-parent-guard](raw/full-0854c1cf41-parent-guard.json) | 0 | 69.213 | 976872 | src/compiler/symbolWalker.ts |
| [full-4dd6a78338-head-guard](raw/full-4dd6a78338-head-guard.json) | 0 | 44.833 | 889900 | src/compiler/transformers/classThis.ts |
| [full-4dd6a78338-parent-base](raw/full-4dd6a78338-parent-base.json) | 0 | 45.861 | 903412 | src/compiler/transformers/classThis.ts |
| [full-4dd6a78338-parent-guard](raw/full-4dd6a78338-parent-guard.json) | 0 | 43.319 | 868732 | src/compiler/transformers/classThis.ts |
| [full-5dc3fc14d8-head-guard](raw/full-5dc3fc14d8-head-guard.json) | 0 | 96.436 | 1011340 | src/compiler/transformers/destructuring.ts |
| [full-5dc3fc14d8-parent-base](raw/full-5dc3fc14d8-parent-base.json) | 0 | 85.591 | 1011288 | src/compiler/transformers/destructuring.ts |
| [full-5dc3fc14d8-parent-guard](raw/full-5dc3fc14d8-parent-guard.json) | 0 | 98.389 | 1015604 | src/compiler/transformers/destructuring.ts |
| [full-7cabb39c93-head-guard](raw/full-7cabb39c93-head-guard.json) | 0 | 44.469 | 867988 | src/compiler/factory/nodeChildren.ts |
| [full-7cabb39c93-parent-base](raw/full-7cabb39c93-parent-base.json) | 0 | 48.937 | 877344 | src/compiler/factory/nodeChildren.ts |
| [full-7cabb39c93-parent-guard](raw/full-7cabb39c93-parent-guard.json) | 0 | 46.055 | 879456 | src/compiler/factory/nodeChildren.ts |
| [full-7f2a045a69-head-guard](raw/full-7f2a045a69-head-guard.json) | 0 | 54.414 | 966836 | src/compiler/transformers/module/impliedNodeFormatDependent.ts |
| [full-7f2a045a69-parent-base](raw/full-7f2a045a69-parent-base.json) | 0 | 48.300 | 931080 | src/compiler/transformers/module/impliedNodeFormatDependent.ts |
| [full-7f2a045a69-parent-guard](raw/full-7f2a045a69-parent-guard.json) | 0 | 50.528 | 963416 | src/compiler/transformers/module/impliedNodeFormatDependent.ts |
| [full-8f10524394-parent-base](raw/full-8f10524394-parent-base.json) | 0 | 94.473 | 916584 | src/compiler/factory/parenthesizerRules.ts |
| [full-91560bd618-head-guard](raw/full-91560bd618-head-guard.json) | 0 | 103.929 | 1051488 | src/compiler/transformers/typeSerializer.ts |
| [full-91560bd618-parent-base](raw/full-91560bd618-parent-base.json) | 0 | 103.060 | 1038416 | src/compiler/transformers/typeSerializer.ts |
| [full-91560bd618-parent-guard](raw/full-91560bd618-parent-guard.json) | 0 | 104.451 | 1073548 | src/compiler/transformers/typeSerializer.ts |
| [full-a48963c320-head-guard](raw/full-a48963c320-head-guard.json) | 0 | 58.402 | 896356 | src/compiler/transformers/namedEvaluation.ts |
| [full-a48963c320-parent-base](raw/full-a48963c320-parent-base.json) | 0 | 49.777 | 871532 | src/compiler/transformers/namedEvaluation.ts |
| [full-a48963c320-parent-guard](raw/full-a48963c320-parent-guard.json) | 0 | 52.175 | 902648 | src/compiler/transformers/namedEvaluation.ts |
| [full-control-head-base](raw/full-control-head-base.json) | 0 | 0.194 | 32576 | /workspace/roads-guard-control/control.ts |
| [full-control-head-mutant](raw/full-control-head-mutant.json) | 0 | 0.229 | 31876 | /workspace/roads-guard-control/control.ts |
| [full-control-head-real](raw/full-control-head-real.json) | 0 | 0.194 | 32876 | /workspace/roads-guard-control/control.ts |
| [full-control-parent-base](raw/full-control-parent-base.json) | 0 | 0.540 | 31804 | /workspace/roads-guard-control/control.ts |
| [full-control-parent-mutant](raw/full-control-parent-mutant.json) | 0 | 0.269 | 31976 | /workspace/roads-guard-control/control.ts |
| [full-control-parent-real](raw/full-control-parent-real.json) | 0 | 0.203 | 32584 | /workspace/roads-guard-control/control.ts |
| [full-d082b22d0d-head-guard](raw/full-d082b22d0d-head-guard.json) | 0 | 127.071 | 1126208 | src/compiler/transformers/module/esnextAnd2015.ts |
| [full-d082b22d0d-parent-base](raw/full-d082b22d0d-parent-base.json) | 0 | 128.424 | 1114504 | src/compiler/transformers/module/esnextAnd2015.ts |
| [full-d082b22d0d-parent-guard](raw/full-d082b22d0d-parent-guard.json) | 0 | 128.994 | 1099408 | src/compiler/transformers/module/esnextAnd2015.ts |
| [full-dbf13ff563-head-guard](raw/full-dbf13ff563-head-guard.json) | 0 | 36.476 | 871208 | src/compiler/factory/utilitiesPublic.ts |
| [full-dbf13ff563-parent-base](raw/full-dbf13ff563-parent-base.json) | 0 | 33.285 | 885928 | src/compiler/factory/utilitiesPublic.ts |
| [full-dbf13ff563-parent-guard](raw/full-dbf13ff563-parent-guard.json) | 0 | 36.193 | 889316 | src/compiler/factory/utilitiesPublic.ts |
| [guard-head](raw/guard-head.json) | 0 | 4911.040 | 4490772 |  |
| [guard-parent](raw/guard-parent.json) | 0 | 4888.397 | 4624656 |  |
| [identity-head](raw/identity-head.json) | 0 | 0.396 | 32468 |  |
| [identity-parent](raw/identity-parent.json) | 0 | 0.291 | 32040 |  |
| [pair-guard-first](raw/pair-guard-first.json) | 0 | 16.125 | 346056 |  |
| [pair-tests](raw/pair-tests.json) | 0 | 52.419 | 82492 |  |
