// cohere/internal/format/doc/width.go, using the same captured Unicode tables and emoji RE2 programs.
import { panic } from 'adamic';
import { emoji, emojiStart, narrow, narrowStart, wideRanges } from './widthTables.ts';
function mapped(code: number): number {
    return code >= 0xd800 && code <= 0xdfff ? 0x100000 + code - 0xd800 : code;
}
// Ordered alternatives match RE2's leftmost-first choice. The patterns have no capturing effects.
function match(program: readonly (readonly number[])[], instruction: number, text: string, position: number): number {
    const row = program[instruction] ?? panic('missing width instruction');
    const op = row[0] ?? panic('missing width operation');
    const out = row[1] ?? panic('missing width successor');
    const arg = row[2] ?? panic('missing width argument');
    if(op === 0 || op === 1) {
        const first = match(program, out, text, position);
        return first >= 0 ? first : match(program, arg, text, position);
    }
    if(op === 2 || op === 6) {
        return match(program, out, text, position);
    }
    if(op === 3) {
        if((arg & 4) !== 0 && position !== 0) {
            return -1;
        }
        if((arg & 8) !== 0 && position !== text.length) {
            return -1;
        }
        return match(program, out, text, position);
    }
    if(op === 4) {
        return position;
    }
    if(op === 5 || position >= text.length) {
        return -1;
    }
    const code = mapped(text.charCodeAt(position));
    let accepted = op === 9 || (op === 10 && code !== 10);
    if(row.length === 4) {
        accepted = code === row[3];
    }
    else {
        for(let index = 3; index + 1 < row.length; index += 2) {
            if(
                code >= (row[index] ?? panic('missing rune start')) &&
                code <= (row[index + 1] ?? panic('missing rune end'))
            ) {
                accepted = true;
                break;
            }
        }
    }
    return accepted ? match(program, out, text, position + 1) : -1;
}
function wide(code: number): boolean {
    let left = 0;
    let right = wideRanges.length;
    while(left < right) {
        const middle = Math.floor((left + right) / 2);
        const range = wideRanges[middle] ?? panic('missing wide range');
        if((range[1] ?? panic('missing wide end')) < code) {
            left = middle + 1;
        }
        else {
            right = middle;
        }
    }
    const range = wideRanges[left];
    return range !== undefined && (range[0] ?? panic('missing wide start')) <= code;
}
// Epsilon paths have no observable captures in this fixed width pattern.
// Keep their depth-first branch order, indexed by the first mapped UTF-16 unit.
function emojiEntries(): (number[] | undefined)[] {
    // Direct UTF-16 indexing avoids the runtime's numeric Map hash probes.
    const result: (number[] | undefined)[] = [];
    for(let unit = 0; unit < 65536; unit++) result.push(undefined);
    const stack: number[] = [emojiStart];
    while(stack.length > 0) {
        const instruction = stack.pop() ?? -1;
        const row = emoji[instruction] ?? panic('missing emoji first instruction');
        const op = row[0] ?? -1;
        if(op === 0 || op === 1) {
            stack.push(row[2] ?? -1);
            stack.push(row[1] ?? -1);
        }
        else if(op === 2 || op === 6) stack.push(row[1] ?? -1);
        else if(op !== 5) {
            // This pinned unanchored pattern always consumes a rune first.
            if(op === 3 || op === 4 || op === 9 || op === 10) panic('unexpected emoji first operation');
            for(let index = 3; index < row.length; index += row.length === 4 ? 1 : 2) {
                const first = row[index] ?? 0;
                const last = row.length === 4 ? first : (row[index + 1] ?? first);
                for(let code = first; code <= last; code++) {
                    const unit = code >= 0x100000 ? code - 0x100000 + 0xd800 : code;
                    const entries = result[unit];
                    if(entries === undefined) result[unit] = [instruction];
                    else entries.push(instruction);
                }
            }
        }
    }
    return result;
}
const emojiFirst = emojiEntries();
function matchEmoji(text: string, position: number): number {
    const entries = emojiFirst[text.charCodeAt(position)];
    if(entries === undefined) return -1;
    for(const instruction of entries) {
        const end = match(emoji, instruction, text, position);
        if(end >= 0) return end;
    }
    return -1;
}
export function stringWidth(text: string): number {
    let ascii = true;
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        if(code < 32 || code > 127) {
            ascii = false;
            break;
        }
    }
    if(ascii) {
        return text.length;
    }
    let width = 0;
    for(let index = 0; index < text.length;) {
        const end = matchEmoji(text, index);
        if(end > index) {
            const found = text.slice(index, end);
            width += match(narrow, narrowStart, found, 0) === found.length ? 1 : 2;
            index = end;
            continue;
        }
        const code = text.codePointAt(index) ?? panic('missing width point');
        index += code > 0xffff ? 2 : 1;
        if(
            code <= 31 ||
            (code >= 127 && code <= 159) ||
            (code >= 0x300 && code <= 0x36f) ||
            (code >= 0xfe00 && code <= 0xfe0f)
        ) {
            continue;
        }
        width += wide(code) ? 2 : 1;
    }
    return width;
}
