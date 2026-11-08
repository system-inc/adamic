# Branded call census checkpoint

Compiler parent: b5c0aa648a7a35b2f17899895d009eabd4147000.
Compiler head: 351fc63fa2573036e194656fb92c35fe03858afe.
TypeScript source: 050880ce59e30b356b686bd3144efe24f875ebc8.

Pairing A uses the parent's adaptations. Pairing B adds only the two requested
adaptation directories from eeb87728b743235cf1df68dfed17241f63416734.
Applies and setup are in progress. No census result is claimed yet.

Checkpoints use separate suffixed branch refs so the final requested branch can
contain exactly one commit directly on b5c0aa64 without any force push.

Initial A apply failed with exit 1 after 425.175 seconds: its
20-optional-declarations Node child received SIGKILL during concurrent setup,
applies and compiler builds. This is an infrastructure failure, not a compiler
regression. The two census builds were then terminated to reduce memory demand.
A will be retried from a fresh adapted output after heavier setup work finishes.

Pairing B apply completed successfully. Requested checker coordinates shift by 3 or 4 lines in this exact applied tree; the AST inventory retains requested and observed coordinates. Measurements are pending.

Pairing A retry apply completed successfully, exit 0, 572.863 seconds. Both adapted trees are ready.
