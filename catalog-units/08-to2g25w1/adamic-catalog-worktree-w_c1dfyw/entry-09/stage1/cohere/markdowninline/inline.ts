// Parser-free leaf printers from cohere/internal/format/markdown/{word,print}.go.
// Offsets and flank characters are UTF-16 units, as in the original non-u expression.
import { panic } from 'adamic';
import { punctuation } from './classes.ts';

type MatchType = { preceding: number; run: number; following: number; matched: boolean };
function lineTerminator(code: number): boolean {
    return code === 10 || code === 13 || code === 0x2028 || code === 0x2029;
}
function rest(text: string, start: number): MatchType {
    const marker = text.charCodeAt(start);
    if(marker !== 42 && marker !== 95) return { preceding: start, run: 0, following: 0, matched: false };
    let end = start;
    while(end < text.length && text.charCodeAt(end) === marker) end++;
    for(; end > start; end--) {
        if(end === text.length || !lineTerminator(text.charCodeAt(end))) {
            return { preceding: start, run: end, following: end === text.length ? end : end + 1, matched: true };
        }
    }
    return { preceding: start, run: 0, following: 0, matched: false };
}
function match(text: string, position: number): MatchType {
    let end = position;
    while(end < text.length && text.charCodeAt(end) === 92) end++;
    for(; end > position; end--) {
        const answer = rest(text, end);
        if(answer.matched) return answer;
    }
    if(position === 0) {
        const answer = rest(text, 0);
        if(answer.matched) return answer;
    }
    if(!lineTerminator(text.charCodeAt(position))) return rest(text, position + 1);
    return { preceding: 0, run: 0, following: 0, matched: false };
}
function whitespace(code: number): boolean {
    return (
        code === 9 ||
        code === 10 ||
        code === 12 ||
        code === 13 ||
        code === 32 ||
        code === 0xa0 ||
        code === 0x1680 ||
        (code >= 0x2000 && code <= 0x200a) ||
        code === 0x202f ||
        code === 0x205f ||
        code === 0x3000
    );
}
function canOpenOrClose(preceding: number, marker: number, following: number): boolean {
    if(preceding < 0 || following < 0) return false;
    const beforeSpace = whitespace(preceding);
    const afterSpace = whitespace(following);
    const beforePunctuation = punctuation(preceding);
    const afterPunctuation = punctuation(following);
    const left = !afterSpace && (!afterPunctuation || beforeSpace || beforePunctuation);
    const right = !beforeSpace && (!beforePunctuation || afterSpace || afterPunctuation);
    if(marker === 42) return left || right;
    if(left) return !right || beforePunctuation;
    if(right) return !left || afterPunctuation;
    return false;
}
export function escapeDelimiterRuns(text: string, previous: string, next: string): string {
    if(!text.includes('*') && !text.includes('_')) return text;
    const parts: string[] = [];
    let start = 0;
    let position = 0;
    while(position < text.length) {
        const found = match(text, position);
        if(!found.matched) {
            position++;
            continue;
        }
        let backslashes = true;
        for(let index = position; index < found.preceding; index++) {
            if(text.charCodeAt(index) !== 92) backslashes = false;
        }
        const alreadyEscaped = backslashes && (found.preceding - position) % 2 === 1;
        const before =
            found.preceding > position
                ? text.charCodeAt(found.preceding - 1)
                : previous.length > 0
                  ? previous.charCodeAt(previous.length - 1)
                  : -1;
        const after =
            found.following > found.run ? text.charCodeAt(found.run) : next.length > 0 ? next.charCodeAt(0) : -1;
        if(!alreadyEscaped && canOpenOrClose(before, text.charCodeAt(found.preceding), after)) {
            // Cut only at ASCII delimiter boundaries, never between a surrogate pair.
            parts.push(text.slice(start, found.preceding));
            parts.push('\\');
            start = found.preceding;
        }
        position = found.following;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
export function printWord(
    value: string,
    emphasis: boolean,
    leading: boolean,
    pseudoSetext: boolean,
    previous: string,
    next: string,
): string {
    let text = value;
    if(!emphasis) {
        if(pseudoSetext && text !== '') {
            const marker = text.charCodeAt(0);
            let fake = marker === 61 || marker === 45;
            for(let index = 1; index < text.length; index++) {
                if(text.charCodeAt(index) !== marker) fake = false;
            }
            if(fake) return `\\${text}`;
        }
        return text;
    }
    if(leading && (text.startsWith('*') || text.startsWith('_'))) text = `\\${text}`;
    return escapeDelimiterRuns(text, previous, next);
}
function replaceCharacter(text: string, character: number, replacement: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.charCodeAt(index) === character) {
            parts.push(text.slice(start, index));
            parts.push(replacement);
            start = index + 1;
        }
    }
    parts.push(text.slice(start));
    return parts.join('');
}
export function printInlineCode(value: string, preserve: boolean, table: boolean): string {
    let code = preserve ? value : replaceCharacter(value, 10, ' ');
    if(table) code = replaceCharacter(code, 124, '\\|');
    const runs: number[] = [];
    let run = 0;
    for(let index = 0; index <= code.length; index++) {
        if(code.charCodeAt(index) === 96) run++;
        else if(run > 0) {
            runs.push(run);
            run = 0;
        }
    }
    let count = 1;
    while(runs.includes(count)) count++;
    const fence = '`'.repeat(count);
    let nonSpace = false;
    for(let index = 0; index < code.length; index++) {
        if(code.charCodeAt(index) !== 10 && code.charCodeAt(index) !== 32) nonSpace = true;
    }
    const first = code.charCodeAt(0);
    const last = code.charCodeAt(code.length - 1);
    const padding =
        first === 96 || last === 96 || ((first === 10 || first === 32) && (last === 10 || last === 32) && nonSpace)
            ? ' '
            : '';
    return fence + padding + code + padding + fence;
}
export function printWikiLink(value: string, preserve: boolean): string {
    if(preserve) return `[[${value}]]`;
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < value.length; index++) {
        const code = value.charCodeAt(index);
        if(code !== 9 && code !== 10) continue;
        parts.push(value.slice(start, index));
        while(index + 1 < value.length && (value.charCodeAt(index + 1) === 9 || value.charCodeAt(index + 1) === 10))
            index++;
        parts.push(' ');
        start = index + 1;
    }
    parts.push(value.slice(start));
    return `[[${parts.join('')}]]`;
}
export function printURL(url: string, dangerous: string): string {
    if(!url.includes(' ') && (dangerous === '' || !url.includes(dangerous))) return url;
    return `<${replaceCharacter(replaceCharacter(url, 60, '%3C'), 62, '%3E')}>`;
}
export function printTitle(title: string, single: boolean): string {
    if(title === '') return '';
    if(title.includes('"') && title.includes("'") && !title.includes(')')) return `(${title})`;
    let preferred = 0;
    let alternate = 0;
    const quote = single ? "'" : '"';
    const other = single ? '"' : "'";
    for(let index = 0; index < title.length; index++) {
        const character = title.slice(index, index + 1);
        if(character === quote) preferred++;
        if(character === other) alternate++;
    }
    const chosen = preferred > alternate ? other : quote;
    return chosen + replaceCharacter(replaceCharacter(title, 92, '\\\\'), chosen.charCodeAt(0), `\\${chosen}`) + chosen;
}
export function printReference(label: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < label.length; index++) {
        const code = label.charCodeAt(index);
        const space = whitespace(code) || code === 11 || code === 0x2028 || code === 0x2029 || code === 0xfeff;
        if(space || code === 92 || code === 91 || code === 93) {
            parts.push(label.slice(start, index));
            if(space) {
                while(index + 1 < label.length) {
                    const next = label.charCodeAt(index + 1);
                    if(!whitespace(next) && next !== 11 && next !== 0x2028 && next !== 0x2029 && next !== 0xfeff) break;
                    index++;
                }
                parts.push(' ');
            }
            else parts.push(`\\${label.slice(index, index + 1)}`);
            start = index + 1;
        }
    }
    parts.push(label.slice(start));
    return `[${parts.join('')}]`;
}
export function printImage(alt: string, original: string, url: string, title: string, single: boolean): string {
    const destination = url === '' ? '<>' : printURL(url, ')');
    const printedTitle = printTitle(title, single);
    return `![${original !== '' ? original : alt}](${destination}${printedTitle === '' ? '' : ` ${printedTitle}`})`;
}
// These modes describe leaf nodes and explicit ancestor contexts, not a Markdown parser.
export function formatLeaf(mode: string, text: string): string {
    if(mode === 'w') return printWord(text, true, false, false, 'a', 'a');
    if(mode === 'e') return printWord(text, true, false, false, ' ', ' ');
    if(mode === 'f') return printWord(text, true, true, false, '', '');
    if(mode === 'n') return printWord(text, true, false, false, '', '');
    if(mode === 's') return printWord(text, false, false, true, '\n', '\n');
    if(mode === 'p') return printWord(text, false, false, false, '', '');
    if(mode === 'c' || mode === 't' || mode === 'r' || mode === 'u')
        return printInlineCode(text, mode === 'c' || mode === 't', mode === 't' || mode === 'u');
    if(mode === 'k' || mode === 'v') return printWikiLink(text, mode === 'k');
    if(mode === 'h' || mode === 'i') return printImage(text, '', text, text, mode === 'i');
    if(mode === 'o') return printImage('fallback', text, text, '', false);
    if(mode === 'q') return `!${printReference(text)}`;
    if(mode === 'j') return `![${text}]${printReference(text)}`;
    if(mode === 'l') return `!${printReference(text)}[]`;
    if(mode === 'b') return `[^${text}]`;
    return panic(`unknown leaf mode: ${mode}`);
}
