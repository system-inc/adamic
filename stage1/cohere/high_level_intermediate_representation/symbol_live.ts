import { panic } from 'adamic';
import type { Checker } from '../lint/checker.a';
import { SymbolGraph } from './export_origin.ts';
import { SymbolFacts, SymbolSnapshot } from './symbol.ts';

export function readSymbol(checker: Checker, node: number): SymbolFacts {
    const answer = checker.ask(node, 'symbol');
    return new SymbolFacts(answer.value ?? panic(`HIR symbol query refused: ${answer.reason}`));
}


// Snapshot compiler symbol facts only; no graph answers enter lowering.
export function readSymbols(checker: Checker): SymbolSnapshot {
    const snapshot = new SymbolSnapshot('1\n0');
    for(let index = 0; index < checker.parser.nodes.length; index++) {
        const node = checker.parser.node(index);
        if(node.kind === 'Identifier') { snapshot.symbols.set(`${checker.offsets[node.pos]}:${checker.offsets[node.end]}`, readSymbol(checker, index)); }
    }
    const graph = checker.ask(checker.root, 'symbol-graph');
    snapshot.graph = new SymbolGraph(graph.value ?? panic(`symbol graph query refused: ${graph.reason}`));
    return snapshot;
}
