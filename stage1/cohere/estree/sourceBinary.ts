import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { precedence } from './sourceGrammar.ts';
export function assertionRank(nodes: readonly ParseNode[], last: number): number {
    const previous = nodes[last];
    return previous?.kind === 'BinaryExpression' ? precedence(nodes[previous.children[1] ?? -1]?.kind ?? '') : 99;
}
export function bindsBeforeAssertion(operator: string, lastRank: number): boolean {
    const nextRank = precedence(operator);
    if(nextRank > lastRank || (nextRank === lastRank && operator === 'AsteriskAsteriskToken')) {
        return true;
    }
    return false;
}
