// Port of pinned Go cohere's no-empty-character-class.
import type { RuleContext } from './rule_context.ts';

const messageEmptyCharacterClass =
    'This character class is empty, so it matches nothing and makes the whole pattern unmatchable. Nobody writes an unmatchable pattern on purpose, which is why this is almost always a class whose contents were deleted or never typed. It is worth reporting because the failure is silent: the regex compiles, the code runs, and the branch behind the match simply never fires.';

function escapeEnd(pattern: string, start: number, unicode: boolean): number {
    if(start + 1 >= pattern.length) return -1;
    const letter = pattern[start + 1] ?? '';
    if(unicode && ['u', 'p', 'P', 'q'].includes(letter) && pattern[start + 2] === '{') {
        const end = pattern.indexOf('}', start + 3);
        if(end >= 0) return end + 1;
    }
    if(letter === 'c' && start + 2 < pattern.length) return start + 3;
    const width = letter === 'x' ? 2 : letter === 'u' ? 4 : 0;
    if(width > 0 && start + 2 + width <= pattern.length) {
        let valid = true;
        for(let offset = 0; offset < width; offset++) {
            if(!'0123456789abcdefABCDEF'.includes(pattern[start + 2 + offset] ?? '')) valid = false;
        }
        if(valid) return start + 2 + width;
    }
    return start + ((pattern.codePointAt(start + 1) ?? 0) > 65535 ? 3 : 2);
}

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-empty-character-class')) return;
    const node = ctx.node(index);
    if(node.kind !== 'RegularExpressionLiteral') return;
    const slash = node.text.lastIndexOf('/');
    const pattern = node.text.slice(1, slash);
    const flags = node.text.slice(slash + 1);
    const nested = flags.includes('v');
    const unicode = nested || flags.includes('u');
    const stack: number[] = [];
    let empty: number[] = [];
    for(let cursor = 0; cursor < pattern.length;) {
        const character = pattern[cursor];
        if(character === '\\') {
            cursor = escapeEnd(pattern, cursor, unicode);
            if(cursor < 0) return;
            continue;
        }
        if(character === '[' && (nested || stack.length === 0)) stack.push(cursor);
        if(character === ']' && stack.length > 0) {
            const opening = stack.pop() ?? -1;
            if(cursor === opening + 1) empty.push(opening);
            if(stack.length === 0) {
                for(const position of empty) {
                    const start = ctx.start(index) + 1 + position;
                    const finding = ctx.report(
                        index,
                        'no-empty-character-class',
                        'unexpectedEmptyCharacterClass',
                        messageEmptyCharacterClass,
                    );
                    finding.editStart = start;
                    finding.editEnd = start + 2;
                    ctx.replaceRange(finding, start, start + 2);
                }
                empty = [];
            }
        }
        cursor++;
    }
}
