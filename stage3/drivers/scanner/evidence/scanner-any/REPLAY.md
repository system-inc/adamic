# Native discovery measurement

These are throwing-placeholder checkpoints, separate from the real scanner
runs. Compiler source is cloud/land-area-next 28285421 without records/library.
The real, uncommitted scanner tree has permanent 42 and the README's existing
slice feature adaptations. Native split 0/1 stop at the original MapLike gate.

`initial-checkpoint.patch.gz` reconstructs the checkpoint immediately before
the private-any site from that 42 slice, using LF text. It includes the preceding
unit's unowned placeholders in corePublic, Debug, diagnostics, utilities and
core, and only getIdentifierToken/UTF16-worker placeholders in scanner.ts.
The private error forwarding body, map declaration and actual driver callback
are intact. The concrete control reaches string/number concatenation; replacing
only both private arg0 annotations with any restores 608:87.

`run-checkpoints.py` records the exact one-shot commands used after preparing
that initial checkpoint in `/workspace/scratch/scanner-any-sites`. It then:

1. Throws from the whole createScanner factory, preserving its signature.
2. Replaces the computed-key keyword object with a typed throwing initializer.
3. Builds the explicit keyword map, then restores inference and builds again.
4. Replaces the two unrelated unsupported map initializers with typed throwers.
5. Builds the actual concrete driver callback, then removes its annotation.

The map comparison is dependent on the empty MapLike discovery placeholder.
Explicit types eliminate Map<any>, but its Object.entries still lacks the pairs
lowering expects. This does not attribute a production Map<any> to pristine tsc.
The driver reaches union JSON.stringify; its restored contextual any fails sooner.
`discovery-scanner.patch.gz` retains the final scanner placeholders. Neither
placeholder patch is a source adaptation or a native correctness claim.

Small `.a` witnesses independently build and match Node for each concrete type;
restored-any source mutants refuse. `run-witnesses.py` records their commands,
including transpiling each actual mutant for Node and planting exactly one
changed byte in each successful native control's output. Their unused payloads
prove type admission, not the scanner's payload runtime representation.
No discovery source tree or compiler implementation was committed or pushed.
