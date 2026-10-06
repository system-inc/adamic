// Go flag's parsing order, help layout and 64-bit integer values. Integers stay decimal text so
// accepting an int64 never rounds it through a JavaScript number.
import { definitions, renameDefinitions } from './definitions.ts';
import { quoted } from '../config/json.ts';

export class Arguments {
    readonly values = new Map<string, string>();
    readonly given = new Map<string, boolean>();
    names: string[] = [];
    stderr = '';
    exitCode = 0;
    stopped = false;
    constructor(renaming: boolean) {
        for(const item of renaming ? renameDefinitions : definitions) {
            this.values.set(item.name, item.value);
        }
    }
    value(name: string): string {
        return this.values.get(name) ?? '';
    }
    enabled(name: string): boolean {
        return this.value(name) === 'true';
    }
}

export function help(program: string, renaming = false): string {
    let text = renaming
        ? 'usage: cohere rename <file>:<line>:<column> <newName>\n       cohere rename <name> <newName>\n\nRenames a symbol everywhere the type graph says it is used.\nThe position form is primary; a bare name is refused when it is ambiguous.\nNothing is written without --write.\n\n'
        : program === ''
          ? 'Usage:\n'
          : `Usage of ${program}:\n`;
    for(const item of renaming ? renameDefinitions : definitions) {
        let parameter = item.kind === 'Bool' ? '' : item.kind === 'Int' ? 'int' : 'string';
        let usage = item.usage;
        const opening = usage.indexOf('`');
        const closing = opening < 0 ? -1 : usage.indexOf('`', opening + 1);
        if(closing >= 0) {
            parameter = usage.slice(opening + 1, closing);
            usage = usage.slice(0, opening) + parameter + usage.slice(closing + 1);
        }
        text += `  -${item.name}${parameter === '' ? '' : ' ' + parameter}\n    \t${usage}`;
        if(item.value !== '' && item.value !== 'false' && item.value !== '0') {
            text += ` (default ${item.kind === 'String' ? quoted(item.value) : item.value})`;
        }
        text += '\n';
    }
    return text;
}

// Return normalized int64 text, or a prefixed parse error. Digit arithmetic remains bounded by 169.
export function integer(input: string): string {
    let text = input;
    let negative = false;
    if(text.startsWith('-') || text.startsWith('+')) {
        negative = text.startsWith('-');
        text = text.slice(1);
    }
    let radix = 10;
    let prefixed = false;
    if(text.startsWith('0x') || text.startsWith('0X')) {
        radix = 16;
        prefixed = true;
        text = text.slice(2);
    }
    else if(text.startsWith('0b') || text.startsWith('0B')) {
        radix = 2;
        prefixed = true;
        text = text.slice(2);
    }
    else if(text.startsWith('0o') || text.startsWith('0O')) {
        radix = 8;
        prefixed = true;
        text = text.slice(2);
    }
    else if(text.length > 1 && text.startsWith('0')) {
        radix = 8;
        prefixed = true;
        text = text.slice(1);
    }
    if(text === '') {
        return '!parse error';
    }
    let decimal = '0';
    let previousDigit = prefixed;
    let digits = 0;
    let underscoresValid = true;

    for(const character of text) {
        if(character === '_') {
            if(!previousDigit) {
                underscoresValid = false;
            }
            previousDigit = false;
            continue;
        }
        const digit = '0123456789abcdef'.indexOf(character.toLowerCase());
        if(digit < 0 || digit >= radix) {
            return '!parse error';
        }
        previousDigit = true;
        digits++;
        let carry = digit;
        let result = '';
        for(let index = decimal.length - 1; index >= 0; index--) {
            const value = (decimal.charCodeAt(index) - 48) * radix + carry;
            result = String.fromCharCode(48 + (value % 10)) + result;
            carry = Math.floor(value / 10);
        }
        while(carry > 0) {
            result = String.fromCharCode(48 + (carry % 10)) + result;
            carry = Math.floor(carry / 10);
        }
        while(result.length > 1 && result.startsWith('0')) {
            result = result.slice(1);
        }
        decimal = result;
        const unsignedLimit = '18446744073709551615';
        if(
            decimal.length > unsignedLimit.length ||
            (decimal.length === unsignedLimit.length && decimal > unsignedLimit)
        ) {
            return '!value out of range';
        }
    }
    if(!previousDigit || !underscoresValid || digits === 0) {
        return '!parse error';
    }
    const limit = negative ? '9223372036854775808' : '9223372036854775807';
    if(decimal.length > limit.length || (decimal.length === limit.length && decimal > limit)) {
        return '!value out of range';
    }
    return negative && decimal !== '0' ? '-' + decimal : decimal;
}

export function parseArguments(program: string, input: string[], renaming = false): Arguments {
    const result = new Arguments(renaming);
    let at = 0;
    let error = '';
    while(at < input.length) {
        const token = input[at] ?? '';
        if(token.length < 2 || !token.startsWith('-')) {
            break;
        }
        at++;
        if(token === '--') {
            break;
        }
        let name = token.slice(token.startsWith('--') ? 2 : 1);
        if(name === '' || name.startsWith('-') || name.startsWith('=')) {
            error = `bad flag syntax: ${token}`;
            break;
        }
        const equals = name.indexOf('=');
        let value = equals < 0 ? '' : name.slice(equals + 1);
        if(equals >= 0) {
            name = name.slice(0, equals);
        }
        let kind = '';
        for(const item of renaming ? renameDefinitions : definitions) {
            if(item.name === name) {
                kind = item.kind;
                break;
            }
        }
        if(kind === '') {
            if(name === 'h' || name === 'help') {
                result.stderr = help(program, renaming);
                result.stopped = true;
                result.names = input.slice(at);
                if(renaming) {
                    result.exitCode = 1;
                    result.stderr += 'cohere: flag: help requested\n';
                }
                return result;
            }
            error = `flag provided but not defined: -${name}`;
            break;
        }
        if(kind === 'Bool') {
            if(equals < 0) {
                value = 'true';
            }
            if(['1', 't', 'T', 'TRUE', 'true', 'True'].includes(value)) {
                value = 'true';
            }
            else if(['0', 'f', 'F', 'FALSE', 'false', 'False'].includes(value)) {
                value = 'false';
            }
            else {
                result.values.set(name, 'false');
                error = `invalid boolean value ${quoted(value)} for -${name}: parse error`;
                break;
            }
        }
        else {
            if(equals < 0) {
                if(at >= input.length) {
                    error = `flag needs an argument: -${name}`;
                    break;
                }
                value = input[at] ?? '';
                at++;
            }
            if(kind === 'Int') {
                const converted = integer(value);
                if(converted.startsWith('!')) {
                    result.values.set(
                        name,
                        converted === '!parse error'
                            ? '0'
                            : value.startsWith('-')
                              ? '-9223372036854775808'
                              : '9223372036854775807',
                    );
                    error = `invalid value ${quoted(value)} for flag -${name}: ${converted.slice(1)}`;
                    break;
                }
                value = converted;
            }
        }
        result.values.set(name, value);
        result.given.set(name, true);
    }
    result.names = input.slice(at);
    if(error !== '') {
        result.stderr = error + '\n' + help(program, renaming) + (renaming ? `cohere: ${error}\n` : '');
        result.exitCode = renaming ? 1 : 2;
        result.stopped = true;
    }
    return result;
}
