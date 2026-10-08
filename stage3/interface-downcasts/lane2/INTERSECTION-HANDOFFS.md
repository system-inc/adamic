Lane 7 intersection handoffs

Owner: lane 7, codex/views-intersections. This list records boundaries and does not add lane 2 certification credit.

SourceFile.amdDependencies has three original candidate reads. The complete original SourceFile cast followed by amdDependencies[0]!.path prints a on Node, then lowering refuses: `Adamic 0.1 refuses checked view read of field path with unsupported intersection contract; prove or implement the intersection contract before reading this field`. TestCheckedViewRanked24AmdDependencyFrontier pins the boundary. Its cause remains to be investigated; the diagnostic is not a claim that AmdDependency.path itself is declared as an intersection.

Original array-read spans:

- src/compiler/transformers/module/module.ts:567 (24222..24242): `node.amdDependencies`
- src/compiler/emitter.ts:4318 (182602..182635): `currentSourceFile.amdDependencies`
- src/compiler/emitter.ts:4319 (182670..182703): `currentSourceFile.amdDependencies`

Lane 2 continues with independent ranked pairs.
