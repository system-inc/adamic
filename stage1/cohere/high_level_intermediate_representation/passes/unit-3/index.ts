// Unit 3 public pass surface; shared records and index classes remain owner imports.
export { outlineFunctions } from './outline_functions.ts';
export { dropManualMemoization } from './drop_manual_memoization.ts';
export type { ManualMemoizationInterface } from './drop_manual_memoization.ts';
export { inlineIIFE } from './inline_iife.ts';
export { copyNestedBodyInto } from './inline_remap.ts';
export type { InlineRemapInterface } from './inline_remap.ts';
export { collectAssumedInvokedFunctions } from './invoked_functions.ts';
export { Unit3Facts } from './facts.ts';
export { eliminateDeadCode } from './dead_code_elimination.ts';
export type { DeadCodeEliminationResultInterface } from './dead_code_elimination.ts';
export { mergeConsecutiveBlocks } from './merge_consecutive_blocks.ts';
export { forFunctionWithoutManualMemoization } from './memo_cache.ts';
