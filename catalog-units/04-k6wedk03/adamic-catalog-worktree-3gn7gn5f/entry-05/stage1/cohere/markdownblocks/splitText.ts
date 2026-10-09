// Cohere utilities.go splitText, on source text without RegExp or AST fixtures.
import { panic } from 'adamic';
import { inRanges } from './emojiMatcher.ts';
import { cjkRanges, variationSelectorRanges, punctuationRanges, hangulRanges } from './textClasses.ts';
import { TextTokens, type TextTokenInterface } from './textTokens.ts';
function contains(value: string, ranges: readonly number[]): boolean {
    for(const character of value) {
        if(inRanges(character.codePointAt(0) ?? panic('text code point'), ranges)) return true;
    }
    return false;
}
function nonCJK(tokens: TextTokens, value: string): void {
    if(value === '') return;
    tokens.word(
        value,
        'non-cjk',
        false,
        inRanges(value.charCodeAt(0), punctuationRanges),
        inRanges(value.charCodeAt(value.length - 1), punctuationRanges),
    );
}
function word(tokens: TextTokens, value: string): void {
    let start = 0;
    let index = 0;
    while(index < value.length) {
        const point = value.codePointAt(index) ?? panic('split code point');
        const size = point > 0xffff ? 2 : 1;
        if(!inRanges(point, cjkRanges)) {
            index += size;
            continue;
        }
        const match = index;
        nonCJK(tokens, value.slice(start, match));
        index += size;
        if(index < value.length) {
            const selector = value.codePointAt(index) ?? panic('variation code point');
            if(inRanges(selector, variationSelectorRanges)) index += selector > 0xffff ? 2 : 1;
        }
        const text = value.slice(match, index);
        if(contains(text, punctuationRanges)) tokens.word(text, 'cjk-punctuation', true, true, true);
        else if(contains(text, hangulRanges)) tokens.word(text, 'k-letter', false, false, false);
        else tokens.word(text, 'cj-letter', true, false, false);
        start = index;
    }
    nonCJK(tokens, value.slice(start));
}
export function splitText(text: string): readonly TextTokenInterface[] {
    const tokens = new TextTokens();
    let start = 0;
    let index = 0;
    while(index < text.length) {
        const code = text.charCodeAt(index);
        if(code !== 9 && code !== 10 && code !== 32) {
            index++;
            continue;
        }
        word(tokens, text.slice(start, index));
        let newline = false;
        while(index < text.length) {
            const unit = text.charCodeAt(index);
            if(unit !== 9 && unit !== 10 && unit !== 32) break;
            if(unit === 10) newline = true;
            index++;
        }
        tokens.whitespace(newline ? '\n' : ' ');
        start = index;
    }
    word(tokens, text.slice(start));
    return tokens.values;
}
