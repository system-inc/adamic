// Shared cache owns keying and construction; this lane supplies only its passes.
import { ForFunctionWithoutManualMemoization } from '../../replay/index.ts';
import { HIRFile } from '../../cache.ts';
import type { ConstructedHIR } from '../../core.ts';
import { dropManualMemoization } from './drop_manual_memoization.ts';
import { inlineIIFE } from './inline_iife.ts';
export function forFunctionWithoutManualMemoization(file: HIRFile,index: number): ConstructedHIR | undefined {
 return ForFunctionWithoutManualMemoization(file,index,(graph) => { dropManualMemoization(graph.arena.read(graph.root)); },(graph) => inlineIIFE(graph.arena,graph.arena.read(graph.root),true));
}
