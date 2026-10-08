import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import { SymbolGraph } from './export_origin.ts';
import { SymbolFacts, SymbolSnapshot } from './symbol.ts';
export interface SymbolQueryAnswerInterface { readonly value: string | undefined; readonly reason: string; }
export function readSymbolsFrom(parser: Parser, offsets: readonly number[], root: number, ask: (node: number, question: string) => SymbolQueryAnswerInterface): SymbolSnapshot {
    const snapshot = new SymbolSnapshot('1\n0');
    for(let index = 0; index < parser.nodes.length; index++) {
        const node = parser.node(index);
        if(node.kind === 'Identifier') { const answer = ask(index,'symbol'); snapshot.symbols.set(`${offsets[node.pos]}:${offsets[node.end]}`, new SymbolFacts(answer.value ?? panic(`HIR symbol query refused: ${answer.reason}`))); }
    }
    const graph = ask(root,'symbol-graph');
    snapshot.graph = new SymbolGraph(graph.value ?? panic(`symbol graph query refused: ${graph.reason}`));
    return snapshot;
}
