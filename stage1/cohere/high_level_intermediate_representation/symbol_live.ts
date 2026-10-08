import { panic } from 'adamic';
import type { Checker } from '../lint/checker.a';
import { SymbolFacts } from './symbol.ts';

export function readSymbol(checker: Checker, node: number): SymbolFacts {
    const answer = checker.ask(node, 'symbol');
    return new SymbolFacts(answer.value ?? panic(`HIR symbol query refused: ${answer.reason}`));
}

