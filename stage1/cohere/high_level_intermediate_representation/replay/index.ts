// Public import for pass lanes: ../../replay/index.ts.
export { Checkpoint, IdentityGraph, ReplayFunction, readGraph, writeGraph, readCheckpoint, writeCheckpoint, escapeField, unescapeField, integer } from './checkpoint.ts';
export type { AnchorKind, GraphRow, ReplayInstruction, SidecarRow, ExtraIdentity } from './checkpoint.ts';
export { decodeGraph, readPlace, readTerminal, readInstructionValue } from './decode.ts';
export { decodeCheckpoint, InputFacts } from './facts.ts';
export { encodeCheckpoint } from './encode.ts';
export { Payload } from './payload.ts';
export { FunctionIndex, BlockIndex, InstructionIndex, IdentifierIndex, DeclarationIndex, ScopeIndex, ReactiveIndex, PayloadIndex, DependencyPathIndex, DependencyTreeIndex } from '../../arena/arena_index.a';
export { copyPattern, copyPatternWithRemap, copyInstructionValue, copyInstructionValueWithRemap, copyTerminal, copyTerminalWithRemap, CloneFunction } from '../clone.ts';
export { finalizeHIR } from '../graph.ts';
export { replaceInstructionValue } from '../core.ts';
export type { SourceHandleInterface, ManualMemoRootInterface, ManualMemoDependencyInterface, DependencyPathEntryInterface } from '../core.ts';
export { LowerFunction, ForFunctionWithoutManualMemoization, MayNameManualMemoization } from '../cache.ts';
