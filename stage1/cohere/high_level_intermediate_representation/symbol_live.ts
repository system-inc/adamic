import { panic } from 'adamic';
import type { Checker } from '../lint/checker.a';
import { readSymbolsFrom } from './symbol_query.ts';
import { SymbolFacts, SymbolSnapshot } from './symbol.ts';

export function readSymbol(checker: Checker, node: number): SymbolFacts {
    const answer = checker.ask(node, 'symbol');
    return new SymbolFacts(answer.value ?? panic(`HIR symbol query refused: ${answer.reason}`));
}


// Snapshot compiler symbol facts only; no graph answers enter lowering.
export function readSymbols(checker: Checker): SymbolSnapshot { return readSymbolsFrom(checker.parser,checker.offsets,checker.root,(node,question) => checker.ask(node,question)); }
