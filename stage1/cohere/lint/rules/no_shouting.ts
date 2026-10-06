import { utf8Length } from 'adamic';
import type { RuleContext } from './context.ts';
import type { CommentRange } from './comment_ranges.ts';
import { Finding } from '../finding.ts';
import { policyMessage } from '../volume_messages.ts';
import { word } from '../comments.ts';
import { trim } from './shared.ts';
import { allowedUppercaseTokens, currencyCodes, shoutedTwoLetterWords } from './policy_data.ts';

function upper(character: string): boolean {
    return character !== '' && character >= 'A' && character <= 'Z';
}
function alphaNumeric(character: string): boolean {
    return word(character) && character !== '_';
}
function asciiSpace(character: string): boolean {
    return [' ', '\t', '\r', '\n', '\f'].includes(character);
}
function blank(text: string, keepLines: boolean): string {
    if(!keepLines) {
        return ' '.repeat(utf8Length(text));
    }
    let result = '';
    for(let cursor = 0; cursor < text.length; cursor++) {
        const point = text.codePointAt(cursor) ?? 0;
        result += point === 10 ? '\n' : ' ';
        if(point > 65535) {
            cursor++;
        }
    }
    return result;
}
function paired(text: string, opening: string, closing: string, lineOnly: boolean): string {
    let cursor = 0;
    let result = '';
    for(;;) {
        const start = text.indexOf(opening, cursor);
        if(start < 0) {
            return result + text.slice(cursor);
        }
        const end = text.indexOf(closing, start + opening.length);
        if(end < 0) {
            return result + text.slice(cursor);
        }
        if(lineOnly && text.slice(start, end).includes('\n')) {
            result += text.slice(cursor, start + opening.length);
            cursor = start + opening.length;
            continue;
        }
        result += text.slice(cursor, start) + blank(text.slice(start, end + closing.length), !lineOnly);
        cursor = end + closing.length;
    }
}
function gutter(text: string, block: boolean): string {
    let cursor = 0;
    while(
        cursor < text.length &&
        (block ? text[cursor] === ' ' || text[cursor] === '\t' : asciiSpace(text[cursor] ?? ''))
    ) {
        cursor++;
    }
    if(text[cursor] === '*') {
        cursor++;
        if(block && cursor < text.length && text[cursor] !== ' ' && text[cursor] !== '\t') {
            return text;
        }
        if(
            cursor < text.length &&
            (block ? text[cursor] === ' ' || text[cursor] === '\t' : asciiSpace(text[cursor] ?? ''))
        ) {
            cursor++;
        }
        return text.slice(cursor);
    }
    return block ? text : text.slice(cursor);
}
function quotedSingles(text: string, phrases: boolean): string {
    let result = '';
    let cursor = 0;
    for(let start = 0; start < text.length; start++) {
        if(text[start] !== "'") {
            continue;
        }
        let end = start + 1;
        while(end < text.length) {
            const character = text[end] ?? '';
            const accepted = phrases
                ? upper(character) ||
                  (character >= '0' && character <= '9') ||
                  ['_', ' ', '.', ',', ':', '=', '-', '/'].includes(character)
                : alphaNumeric(character) || ['_', '.', '-', '/'].includes(character);
            if(!accepted) {
                break;
            }
            end++;
        }
        const contents = text.slice(start + 1, end);
        if(text[end] !== "'" || contents === '') {
            continue;
        }
        if(phrases && (contents.length < 2 || !word(contents[0] ?? '') || !word(contents[contents.length - 1] ?? ''))) {
            continue;
        }
        const skip = phrases && (alphaNumeric(text[start - 1] ?? '') || alphaNumeric(text[end + 1] ?? ''));
        if(!skip) {
            result += text.slice(cursor, start) + ' '.repeat(end - start + 1);
            cursor = end + 1;
        }
        start = end;
    }
    return result + text.slice(cursor);
}
function mask(text: string): string {
    let masked = paired(text, '```', '```', false);
    let lines = masked.split('\n');
    let previousBlank = true;
    let insideBlock = false;
    for(let index = 0; index < lines.length; index++) {
        const line = lines[index] ?? '';
        const content = gutter(line, true);
        if(trim(content) === '') {
            previousBlank = true;
            continue;
        }
        insideBlock = (content.startsWith('    ') || content.startsWith('\t')) && (insideBlock || previousBlank);
        previousBlank = false;
        if(insideBlock) {
            lines[index] = ' '.repeat(utf8Length(line));
        }
    }
    masked = paired(lines.join('\n'), '`', '`', false);
    masked = paired(masked, '"', '"', true);
    lines = masked.split('\n');
    for(let index = 0; index + 1 < lines.length; index++) {
        const one = lines[index] ?? '';
        const two = lines[index + 1] ?? '';
        const opening = one.indexOf('"');
        const closing = two.indexOf('"');
        if(opening >= 0 && closing >= 0) {
            lines[index] = one.slice(0, opening) + ' '.repeat(utf8Length(one.slice(opening)));
            lines[index + 1] = ' '.repeat(utf8Length(two.slice(0, closing + 1))) + two.slice(closing + 1);
        }
    }
    masked = quotedSingles(lines.join('\n'), false);
    masked = quotedSingles(masked, true);
    lines = masked.split('\n');
    let example = false;
    for(let index = 0; index < lines.length; index++) {
        const line = lines[index] ?? '';
        const content = gutter(line, false);
        if(content.startsWith('@example') && !word(content[8] ?? '')) {
            example = true;
        }
        else if(content.startsWith('@') && word(content[1] ?? '')) {
            example = false;
        }
        if(example) {
            lines[index] = ' '.repeat(utf8Length(line));
        }
    }
    for(let index = 0; index < lines.length; index++) {
        const line = lines[index] ?? '';
        let content = gutter(line, false);
        let cursor = 0;
        while(cursor < content.length && asciiSpace(content[cursor] ?? '')) {
            cursor++;
        }
        content = content.slice(cursor);
        let command = content.startsWith('$');
        for(const prefix of ['ahra', 'git', 'pnpm', 'npm', 'sqlite3', 'curl']) {
            if(content.startsWith(prefix) && asciiSpace(content[prefix.length] ?? '')) {
                command = true;
            }
        }
        if(content.startsWith('s') && asciiSpace(content[1] ?? '')) {
            cursor = 1;
            while(cursor < content.length && asciiSpace(content[cursor] ?? '')) {
                cursor++;
            }
            if(content[cursor] === 'c' && !word(content[cursor + 1] ?? '')) {
                command = true;
            }
        }
        if(command) {
            lines[index] = ' '.repeat(utf8Length(line));
        }
    }
    return lines.join('\n');
}
export function noShouting(context: RuleContext, ranges: readonly CommentRange[]): void {
    for(const comment of ranges) {
        let adjacent = false;
        for(let index = 1; index < comment.text.length; index++) {
            if(upper(comment.text[index - 1] ?? '') && upper(comment.text[index] ?? '')) {
                adjacent = true;
                break;
            }
        }
        if(!adjacent) {
            continue;
        }
        const text = mask(comment.text);
        const seen = new Set<string>();
        const tokens: string[] = [];
        for(let index = 0; index < text.length; index++) {
            if(!upper(text[index] ?? '') || word(text[index - 1] ?? '')) {
                continue;
            }
            const start = index;
            while(
                index < text.length &&
                (upper(text[index] ?? '') ||
                    ((text[index] ?? '') >= '0' && (text[index] ?? '') <= '9') ||
                    text[index] === '_')
            ) {
                index++;
            }
            const token = text.slice(start, index);
            if(
                word(text[index] ?? '') ||
                seen.has(token) ||
                allowedUppercaseTokens.includes(token) ||
                currencyCodes.includes(token)
            ) {
                continue;
            }
            let letters = true;
            let vowel = false;
            for(const character of token.split('')) {
                if(!upper(character)) {
                    letters = false;
                }
                if('AEIOUY'.includes(character)) {
                    vowel = true;
                }
            }
            if(!letters || token.length < 2 || (token.length === 2 ? !shoutedTwoLetterWords.includes(token) : !vowel)) {
                continue;
            }
            seen.add(token);
            if(!context.settings.list('allow', []).includes(token)) {
                tokens.push(`"${token}"`);
            }
        }
        if(tokens.length > 0) {
            const rule = 'nexus/consistency-no-shouting';
            context.record(
                new Finding(
                    rule,
                    'shoutingInComment',
                    policyMessage(rule, 'shoutingInComment', ['tokens', tokens.slice(0, 4).join(', ')]),
                    comment.start,
                    comment.end,
                    '',
                    '',
                    '',
                ),
            );
        }
    }
}

export function visit(context: RuleContext, index: number): void {
    if(context.node(index).kind === 'SourceFile' && context.enabled('nexus/consistency-no-shouting')) {
        noShouting(context, context.comments);
    }
}
