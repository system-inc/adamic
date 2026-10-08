// yaml 2.9.0 flow and block scalar resolution. All positions are UTF-16 code units.
import type { Token } from './cst.ts';
import type { CSTParser } from './cstParser.ts';
import { ScalarResolution } from './scalarResolution.ts';
import { char, whitespace, foldLines, escapeCode, emptyContent } from './scalarText.ts';
import { NewlineFold } from './newlineFold.ts';
import { BlockLine } from './blockLine.ts';
export class ScalarResolver {
    parser: CSTParser;
    constructor(parser: CSTParser) {
        this.parser = parser;
    }
    /** @mutates result The scalar result collects diagnostics and comments from each resolution step. */
    end(tokens: readonly number[], start: number, reqSpace: boolean, result: ScalarResolution): number {
        let offset = start;
        let hasSpace = false;
        let separator = '';
        for(const index of tokens) {
            const token = this.parser.get(index);
            switch(token.type) {
                case 'space':
                    hasSpace = true;
                    break;
                case 'comment': {
                    if(reqSpace && !hasSpace)
                        result.tokenError(
                            token,
                            'MISSING_CHAR',
                            'Comments must be separated from other tokens by white space characters',
                        );
                    let comment = token.source.slice(1);
                    if(comment === '') comment = ' ';
                    result.comment += result.comment === '' ? comment : separator + comment;
                    separator = '';
                    break;
                }
                case 'newline':
                    if(result.comment !== '') separator += token.source;
                    hasSpace = true;
                    break;
                default:
                    result.tokenError(token, 'UNEXPECTED_TOKEN', `Unexpected ${token.type} at node end`);
            }
            offset += token.source.length;
        }
        return offset;
    }
    /** @mutates result The scalar result collects diagnostics and comments from each resolution step. */
    characterCode(source: string, offset: number, length: number, base: number, result: ScalarResolution): string {
        const digits = source.slice(offset, offset + length);
        let valid = digits.length === length;
        let code = 0;
        for(let index = 0; index < digits.length; index++) {
            const value = '0123456789abcdef'.indexOf(char(digits, index).toLowerCase());
            if(value < 0) valid = false;
            code = code * 16 + value;
        }
        if(valid && code <= 0x10ffff) return String.fromCodePoint(code);
        const raw = source.slice(offset - 2, offset + length);
        result.error(base + offset - 2, 'BAD_DQ_ESCAPE', `Invalid escape sequence ${raw}`);
        return raw;
    }
    /** @mutates result The scalar result collects diagnostics and comments from each resolution step. */
    doubleQuoted(source: string, base: number, result: ScalarResolution): string {
        let value = '';
        for(let index = 1; index < source.length - 1; index++) {
            const character = char(source, index);
            if(character === '\r' && char(source, index + 1) === '\n') continue;
            if(character === '\n') {
                const folded = new NewlineFold(source, index);
                value += folded.value;
                index = folded.offset;
            }
            else if(character === '\\') {
                index++;
                const next = char(source, index);
                const escaped = escapeCode(next);
                if(escaped !== '') value += escaped;
                else if(next === '\n' || (next === '\r' && char(source, index + 1) === '\n')) {
                    if(next === '\r') index++;
                    while(whitespace(char(source, index + 1))) index++;
                }
                else if(next === 'x' || next === 'u' || next === 'U') {
                    const length = next === 'x' ? 2 : next === 'u' ? 4 : 8;
                    value += this.characterCode(source, index + 1, length, base, result);
                    index += length;
                }
                else {
                    const raw = source.slice(index - 1, index + 1);
                    result.error(base + index - 1, 'BAD_DQ_ESCAPE', `Invalid escape sequence ${raw}`);
                    value += raw;
                }
            }
            else if(whitespace(character)) {
                const start = index;
                while(whitespace(char(source, index + 1))) index++;
                const next = char(source, index + 1);
                if(next !== '\n' && !(next === '\r' && char(source, index + 2) === '\n'))
                    value += source.slice(start, index + 1);
            }
            else value += character;
        }
        if(char(source, source.length - 1) !== '"' || source.length === 1)
            result.error(base + source.length, 'MISSING_CHAR', 'Missing closing "quote');
        return value;
    }
    flow(token: Token): ScalarResolution {
        const result = new ScalarResolution();
        switch(token.type) {
            case 'scalar': {
                result.type = 'PLAIN';
                const first = char(token.source, 0);
                let bad = '';
                switch(first) {
                    case '\t':
                        bad = 'a tab character';
                        break;
                    case ',':
                        bad = 'flow indicator character ,';
                        break;
                    case '%':
                        bad = 'directive indicator character %';
                        break;
                    case '|':
                    case '>':
                        bad = `block scalar indicator ${first}`;
                        break;
                    case '@':
                    case '`':
                        bad = `reserved character ${first}`;
                        break;
                }
                if(bad !== '') result.error(token.offset, 'BAD_SCALAR_START', `Plain value cannot start with ${bad}`);
                result.value = foldLines(token.source);
                break;
            }
            case 'single-quoted-scalar': {
                result.type = 'QUOTE_SINGLE';
                if(char(token.source, token.source.length - 1) !== "'" || token.source.length === 1)
                    result.error(token.offset + token.source.length, 'MISSING_CHAR', "Missing closing 'quote");
                const folded = foldLines(token.source.slice(1, -1));
                for(let index = 0; index < folded.length; index++) {
                    const character = char(folded, index);
                    result.value += character;
                    if(character === "'" && char(folded, index + 1) === "'") index++;
                }
                break;
            }
            case 'double-quoted-scalar':
                result.type = 'QUOTE_DOUBLE';
                result.value = this.doubleQuoted(token.source, token.offset, result);
                break;
            default:
                result.tokenError(token, 'UNEXPECTED_TOKEN', `Expected a flow scalar value, but found: ${token.type}`);
                result.range = [token.offset, token.offset + token.source.length, token.offset + token.source.length];
                return result;
        }
        const end = token.offset + token.source.length;
        const nodeEnd = this.end(token.end, end, true, result);
        result.range = [token.offset, end, nodeEnd];
        return result;
    }
    block(token: Token, atRoot: boolean): ScalarResolution {
        const result = new ScalarResolution();
        const first = this.parser.get(token.props[0] ?? -1);
        if(first.type !== 'block-scalar-header') {
            result.tokenError(first, 'IMPOSSIBLE', 'Block scalar header not found');
            result.range = [token.offset, token.offset, token.offset];
            return result;
        }
        const mode = char(first.source, 0);
        let indent = 0;
        let chomp = '';
        let errorOffset = -1;
        for(let index = 1; index < first.source.length; index++) {
            const character = char(first.source, index);
            if(chomp === '' && (character === '-' || character === '+')) chomp = character;
            else {
                const number = '123456789'.indexOf(character) + 1;
                if(indent === 0 && number !== 0) indent = number;
                else if(errorOffset === -1) errorOffset = token.offset + index;
            }
        }
        if(errorOffset !== -1)
            result.error(
                errorOffset,
                'UNEXPECTED_TOKEN',
                `Block scalar header includes extra characters: ${first.source}`,
            );
        let hasSpace = false;
        let length = first.source.length;
        for(let index = 1; index < token.props.length; index++) {
            const property = this.parser.get(token.props[index] ?? -1);
            switch(property.type) {
                case 'space':
                    hasSpace = true;
                    break;
                case 'newline':
                    break;
                case 'comment':
                    if(!hasSpace)
                        result.tokenError(
                            property,
                            'MISSING_CHAR',
                            'Comments must be separated from other tokens by white space characters',
                        );
                    result.comment = property.source.slice(1);
                    break;
                case 'error':
                    result.tokenError(property, 'UNEXPECTED_TOKEN', property.message);
                    break;
                default:
                    result.tokenError(
                        property,
                        'UNEXPECTED_TOKEN',
                        `Unexpected token in block scalar header: ${property.type}`,
                    );
            }
            length += property.source.length;
        }
        result.type = mode === '>' ? 'BLOCK_FOLDED' : 'BLOCK_LITERAL';
        const lines: BlockLine[] = [];
        if(token.source !== '') for(const line of token.source.split('\n')) lines.push(new BlockLine(line));
        let chompStart = lines.length;
        for(let index = lines.length - 1; index >= 0; index--) {
            const line = lines[index];
            if(line === undefined) break;
            if(emptyContent(line.content)) chompStart = index;
            else break;
        }
        const end = token.offset + length + token.source.length;
        result.range = [token.offset, end, end];
        if(chompStart === 0) {
            if(chomp === '+' && lines.length > 0) result.value = '\n'.repeat(Math.max(1, lines.length - 1));
            return result;
        }
        let trimIndent = token.indent + indent;
        let offset = token.offset + length;
        let contentStart = 0;
        for(let index = 0; index < chompStart; index++) {
            const line = lines[index];
            if(line === undefined) break;
            if(emptyContent(line.content)) {
                if(indent === 0 && line.indent.length > trimIndent) trimIndent = line.indent.length;
            }
            else {
                if(line.indent.length < trimIndent)
                    result.error(
                        offset + line.indent.length,
                        'MISSING_CHAR',
                        'Block scalars with more-indented leading empty lines must use an explicit indentation indicator',
                    );
                if(indent === 0) trimIndent = line.indent.length;
                contentStart = index;
                if(trimIndent === 0 && !atRoot)
                    result.error(offset, 'BAD_INDENT', 'Block scalar values in collections must be indented');
                break;
            }
            offset += line.indent.length + line.content.length + 1;
        }
        for(let index = lines.length - 1; index >= chompStart; index--) {
            const line = lines[index];
            if(line === undefined) break;
            if(line.indent.length > trimIndent) chompStart = index + 1;
        }
        let separator = '';
        let previousMoreIndented = false;
        for(let index = 0; index < contentStart; index++) {
            const line = lines[index];
            if(line !== undefined) result.value += `${line.indent.slice(trimIndent)}\n`;
        }
        for(let index = contentStart; index < chompStart; index++) {
            const line = lines[index];
            if(line === undefined) break;
            let lineIndent = line.indent;
            let content = line.content;
            offset += lineIndent.length + content.length + 1;
            const crlf = char(content, content.length - 1) === '\r';
            if(crlf) content = content.slice(0, -1);
            if(content !== '' && lineIndent.length < trimIndent) {
                const origin = indent === 0 ? 'first line' : 'explicit indentation indicator';
                result.error(
                    offset - content.length - (crlf ? 2 : 1),
                    'BAD_INDENT',
                    `Block scalar lines must not be less indented than their ${origin}`,
                );
                lineIndent = '';
            }
            const trimmed = lineIndent.slice(trimIndent);
            if(result.type === 'BLOCK_LITERAL') {
                result.value += separator + trimmed + content;
                separator = '\n';
            }
            else if(lineIndent.length > trimIndent || char(content, 0) === '\t') {
                if(separator === ' ') separator = '\n';
                else if(!previousMoreIndented && separator === '\n') separator = '\n\n';
                result.value += separator + trimmed + content;
                separator = '\n';
                previousMoreIndented = true;
            }
            else if(content === '') {
                if(separator === '\n') result.value += '\n';
                else separator = '\n';
            }
            else {
                result.value += separator + content;
                separator = ' ';
                previousMoreIndented = false;
            }
        }
        if(chomp === '+') {
            for(let index = chompStart; index < lines.length; index++) {
                const line = lines[index];
                if(line !== undefined) result.value += `\n${line.indent.slice(trimIndent)}`;
            }
            if(char(result.value, result.value.length - 1) !== '\n') result.value += '\n';
        }
        else if(chomp !== '-') result.value += '\n';
        return result;
    }
}
