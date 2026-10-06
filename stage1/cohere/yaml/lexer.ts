// Adapted from yaml 2.9.0, Copyright Eemeli Aro, `ISC` license.
function asciiCharacters(): string[] {
    const units: string[] = [];
    for(let code = 0; code < 128; code++) units.push(String.fromCodePoint(code));
    return units;
}
const ascii: readonly string[] = asciiCharacters();
function character(text: string, index: number): string {
    if(index < 0 || index >= text.length || Number.isNaN(index)) return '';
    const code = text.charCodeAt(index);
    return code < 128 ? (ascii[code] ?? '') : text.slice(index, index + 1);
}
function codeUnit(text: string, index: number): number {
    return index < 0 || index >= text.length || Number.isNaN(index) ? -1 : text.charCodeAt(index);
}
function isEmptyCode(unit: number): boolean {
    return unit === -1 || unit === 32 || unit === 10 || unit === 13 || unit === 9;
}

function isFlowIndicatorCode(unit: number): boolean {
    return unit === 44 || unit === 91 || unit === 93 || unit === 123 || unit === 125;
}
function member(set: string, unit: string): boolean {
    return unit !== '' && set.includes(unit);
}
/*
start -> stream

stream
  directive -> line-end -> stream
  indent + line-end -> stream
  [else] -> line-start

line-end
  comment -> line-end
  newline -> .
  input-end -> end

line-start
  doc-start -> doc
  doc-end -> stream
  [else] -> indent -> block-start

block-start
  seq-item-start -> block-start
  explicit-key-start -> block-start
  map-value-start -> block-start
  [else] -> doc

doc
  line-end -> line-start
  spaces -> doc
  anchor -> doc
  tag -> doc
  flow-start -> flow -> doc
  flow-end -> error -> doc
  seq-item-start -> error -> doc
  explicit-key-start -> error -> doc
  map-value-start -> doc
  alias -> doc
  quote-start -> quoted-scalar -> doc
  block-scalar-header -> line-end -> block-scalar(min) -> line-start
  [else] -> plain-scalar(false, min) -> doc

flow
  line-end -> flow
  spaces -> flow
  anchor -> flow
  tag -> flow
  flow-start -> flow -> flow
  flow-end -> .
  seq-item-start -> error -> flow
  explicit-key-start -> flow
  map-value-start -> flow
  alias -> flow
  quote-start -> quoted-scalar -> flow
  comma -> flow
  [else] -> plain-scalar(true, 0) -> flow

quoted-scalar
  quote-end -> .
  [else] -> quoted-scalar

block-scalar(min)
  newline + peek(indent < min) -> .
  [else] -> block-scalar(min)

plain-scalar(is-flow, min)
  scalar-end(is-flow) -> .
  peek(newline + (indent < min)) -> .
  [else] -> plain-scalar(min)
*/
function isEmpty(unit: string): boolean {
    return isEmptyCode(unit === '' ? -1 : unit.charCodeAt(0));
}

const hexDigits = '0123456789ABCDEFabcdef';
const tagChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-#;/?:@&=+$_.!~*'()";
const flowIndicatorChars = ',[]{}';
const invalidAnchorChars = ' ,[]{}\n\r\t';

// Port of eemeli/yaml 2.9.0's lexer, held to Go cohere's UTF-16 token stream.
// Tokens replace yields; eager calls preserve scalar transitions even at end of input.
export class Lexer {
    atEnd = false;
    blockScalarIndent = -1;
    blockScalarKeep = false;
    buffer = '';
    flowKey = false;
    flowLevel = 0;
    indentNext = 0;
    indentValue = 0;
    lineEndPos = -2;
    lineAvailable = false;
    next = '';
    pos = 0;
    tokens: string[] = [];
    atLineEnd(): boolean {
        let index = this.pos;
        let unit = codeUnit(this.buffer, index);
        while(unit === 32 || unit === 9) {
            index += 1;
            unit = codeUnit(this.buffer, index);
        }
        if(unit === -1 || unit === 35 || unit === 10) return true;
        if(unit === 13) return codeUnit(this.buffer, index + 1) === 10;
        return false;
    }
    charAt(count: number): string {
        return character(this.buffer, this.pos + count);
    }
    continueScalar(offset: number): number {
        let unit = codeUnit(this.buffer, offset);
        if(this.indentNext > 0) {
            let indent = 0;
            while(unit === 32) {
                indent += 1;
                unit = codeUnit(this.buffer, indent + offset);
            }
            if(unit === 13) {
                const next = codeUnit(this.buffer, indent + offset + 1);
                if(next === 10 || (next === -1 && !this.atEnd)) return offset + indent + 1;
            }
            return unit === 10 || indent >= this.indentNext || (unit === -1 && !this.atEnd) ? offset + indent : -1;
        }
        if(unit === 45 || unit === 46) {
            const marker = this.buffer.slice(offset, offset + 3);
            if((marker === '---' || marker === '...') && isEmptyCode(codeUnit(this.buffer, offset + 3))) return -1;
        }
        return offset;
    }
    getLine(): string {
        let end = this.lineEndPos;
        if(end === -2 || (end !== -1 && end < this.pos)) {
            end = this.buffer.indexOf('\n', this.pos);
            this.lineEndPos = end;
        }
        this.lineAvailable = end !== -1 || this.atEnd;
        if(end === -1) return this.atEnd ? this.buffer.slice(this.pos) : '';
        if(character(this.buffer, end - 1) === '\r') end -= 1;
        return this.buffer.slice(this.pos, end);
    }
    hasChars(count: number): boolean {
        return this.pos + count <= this.buffer.length;
    }
    setNext(state: string): string {
        this.buffer = this.buffer.slice(this.pos);
        this.pos = 0;
        this.lineEndPos = -2;
        this.next = state;
        return '';
    }
    peek(count: number): string {
        return this.buffer.slice(this.pos, this.pos + count);
    }
    pushCount(count: number): number {
        if(count > 0) {
            this.tokens.push(this.buffer.slice(this.pos, this.pos + count));
            this.pos += count;
            return count;
        }
        return 0;
    }
    pushSpaces(allowTabs: boolean): number {
        let index = this.pos - 1;
        while(true) {
            index++;
            const unit = codeUnit(this.buffer, index);
            if(unit === 32) continue;
            if(allowTabs && unit === 9) continue;
            break;
        }
        const count = index - this.pos;
        if(count > 0) {
            this.tokens.push(this.buffer.slice(this.pos, this.pos + count));
            this.pos = index;
        }
        return count;
    }
    pushNewline(): number {
        const unit = codeUnit(this.buffer, this.pos);
        if(unit === 10) return this.pushCount(1);
        else if(unit === 13 && codeUnit(this.buffer, this.pos + 1) === 10) return this.pushCount(2);
        return 0;
    }
    parseBlockStart(): string {
        const peeked = this.peek(2);
        const firstUnit = character(peeked, 0);
        const first = peeked.charCodeAt(0);
        const nextUnit =
            first >= 0xd800 && first <= 0xdbff && peeked.charCodeAt(1) >= 0xdc00 && peeked.charCodeAt(1) <= 0xdfff
                ? ''
                : character(peeked, 1);
        if(nextUnit === '' && !this.atEnd) return this.setNext('block-start');
        if((firstUnit === '-' || firstUnit === '?' || firstUnit === ':') && isEmpty(nextUnit)) {
            const count = this.pushCount(1) + this.pushSpaces(true);
            this.indentNext = this.indentValue + 1;
            this.indentValue += count;
            return 'block-start';
        }
        return 'doc';
    }
    parseLineStart(): string {
        const unit = this.charAt(0);
        if(unit === '' && !this.atEnd) return this.setNext('line-start');
        if(unit === '-' || unit === '.') {
            if(!this.atEnd && !this.hasChars(4)) return this.setNext('line-start');
            const piece = this.peek(3);
            if((piece === '---' || piece === '...') && isEmpty(this.charAt(3))) {
                this.pushCount(3);
                this.indentValue = 0;
                this.indentNext = 0;
                return piece === '---' ? 'doc' : 'stream';
            }
        }
        this.indentValue = this.pushSpaces(false);
        if(this.indentNext > this.indentValue && !isEmpty(this.charAt(1))) this.indentNext = this.indentValue;
        return this.parseBlockStart();
    }
    parseStream(): string {
        let line = this.getLine();
        if(!this.lineAvailable) return this.setNext('stream');
        if(character(line, 0) === '\ufeff') {
            this.pushCount(1);
            line = line.slice(1);
        }
        if(character(line, 0) === '%') {
            let directiveEnd = line.length;
            let continuationStart = line.indexOf('#');
            while(continuationStart !== -1) {
                const unit = character(line, continuationStart - 1);
                if(unit === ' ' || unit === '\t') {
                    directiveEnd = continuationStart - 1;
                    break;
                }
                else {
                    continuationStart = line.indexOf('#', continuationStart + 1);
                }
            }
            while(true) {
                const unit = character(line, directiveEnd - 1);
                if(unit === ' ' || unit === '\t') directiveEnd -= 1;
                else break;
            }
            const count = this.pushCount(directiveEnd) + this.pushSpaces(true);
            this.pushCount(line.length - count); // possible comment
            // Upstream creates a generator here without iterating it.
            return 'stream';
        }
        if(this.atLineEnd()) {
            const spaces = this.pushSpaces(true);
            this.pushCount(line.length - spaces);
            this.pushNewline();
            return 'stream';
        }
        this.tokens.push('\x02');
        return this.parseLineStart();
    }
    pushToIndex(index: number, allowEmpty: boolean): number {
        const piece = this.buffer.slice(this.pos, index);
        if(piece !== '') {
            this.tokens.push(piece);
            this.pos += piece.length;
            return piece.length;
        }
        if(allowEmpty) this.tokens.push('');
        return 0;
    }
    pushTag(): number {
        if(this.charAt(1) === '<') {
            let index = this.pos + 2;
            let unit = character(this.buffer, index);
            while(!isEmpty(unit) && unit !== '>') {
                index += 1;
                unit = character(this.buffer, index);
            }
            return this.pushToIndex(unit === '>' ? index + 1 : index, false);
        }
        let index = this.pos + 1;
        let unit = character(this.buffer, index);
        while(unit !== '') {
            if(member(tagChars, unit)) {
                index += 1;
                unit = character(this.buffer, index);
            }
            else if(
                unit === '%' &&
                member(hexDigits, character(this.buffer, index + 1)) &&
                member(hexDigits, character(this.buffer, index + 2))
            ) {
                index += 3;
                unit = character(this.buffer, index);
            }
            else break;
        }
        return this.pushToIndex(index, false);
    }
    pushUntil(header: boolean): number {
        let index = this.pos;
        let unit = character(this.buffer, index);
        while(!(header ? isEmpty(unit) || unit === '#' : unit === '' || member(invalidAnchorChars, unit))) {
            index += 1;
            unit = character(this.buffer, index);
        }
        return this.pushToIndex(index, false);
    }
    pushIndicators(): number {
        let count = 0;
        while(true) {
            switch(this.charAt(0)) {
                case '!':
                    count += this.pushTag();
                    count += this.pushSpaces(true);
                    continue;
                case '&':
                    count += this.pushUntil(false);
                    count += this.pushSpaces(true);
                    continue;
                case '-': // this is an error
                case '?': // this is an error outside flow collections
                case ':': {
                    const inFlow = this.flowLevel > 0;
                    const nextUnit = this.charAt(1);
                    if(isEmpty(nextUnit) || (inFlow && member(flowIndicatorChars, nextUnit))) {
                        if(!inFlow) this.indentNext = this.indentValue + 1;
                        else this.flowKey = false;
                        count += this.pushCount(1);
                        count += this.pushSpaces(true);
                        continue;
                    }
                }
            }
            break;
        }
        return count;
    }
    parseBlockScalarHeader(): number {
        this.blockScalarIndent = -1;
        this.blockScalarKeep = false;
        let index = this.pos;
        while(true) {
            index += 1;
            const unit = codeUnit(this.buffer, index);
            if(unit === 43) this.blockScalarKeep = true;
            else if(unit > 48 && unit <= 57) this.blockScalarIndent = unit - 49;
            else if(unit !== 45) break;
        }
        return this.pushUntil(true);
    }
    parseQuotedScalar(): string {
        const quote = this.charAt(0);
        let end = this.buffer.indexOf(quote, this.pos + 1);
        if(quote === "'") {
            while(end !== -1 && character(this.buffer, end + 1) === "'") end = this.buffer.indexOf("'", end + 2);
        }
        else {
            // double-quote
            while(end !== -1) {
                let count = 0;
                while(character(this.buffer, end - 1 - count) === '\\') count += 1;
                if(count % 2 === 0) break;
                end = this.buffer.indexOf('"', end + 1);
            }
        }
        // Only looking for newlines within the quotes
        const quotedBuffer = this.buffer.slice(0, Math.max(0, end));
        let newline = quotedBuffer.indexOf('\n', this.pos);
        if(newline !== -1) {
            while(newline !== -1) {
                const continuationStart = this.continueScalar(newline + 1);
                if(continuationStart === -1) break;
                newline = quotedBuffer.indexOf('\n', continuationStart);
            }
            if(newline !== -1) {
                // this is an error caused by an unexpected unindent
                end = newline - (character(quotedBuffer, newline - 1) === '\r' ? 2 : 1);
            }
        }
        if(end === -1) {
            if(!this.atEnd) return this.setNext('quoted-scalar');
            end = this.buffer.length;
        }
        this.pushToIndex(end + 1, false);
        return this.flowLevel !== 0 ? 'flow' : 'doc';
    }
    parseBlockScalar(): string {
        let newline = this.pos - 1; // may be -1 if this.pos === 0
        let indent = 0;
        let unit = -1;
        let scanning = true;
        for(let index = this.pos; scanning; index++) {
            unit = codeUnit(this.buffer, index);
            if(unit === -1) break;
            switch(unit) {
                case 32:
                    indent += 1;
                    break;
                case 10:
                    newline = index;
                    indent = 0;
                    break;
                case 13: {
                    const next = codeUnit(this.buffer, index + 1);
                    if(next === -1 && !this.atEnd) return this.setNext('block-scalar');
                    if(next === 10) break;
                    scanning = false;
                    break;
                }
                default:
                    scanning = false;
                    break;
            }
        }
        if(unit === -1 && !this.atEnd) return this.setNext('block-scalar');
        if(indent >= this.indentNext) {
            if(this.blockScalarIndent === -1) this.indentNext = indent;
            else {
                this.indentNext = this.blockScalarIndent + (this.indentNext === 0 ? 1 : this.indentNext);
            }
            do {
                const continuationStart = this.continueScalar(newline + 1);
                if(continuationStart === -1) break;
                newline = this.buffer.indexOf('\n', continuationStart);
            } while(newline !== -1);
            if(newline === -1) {
                if(!this.atEnd) return this.setNext('block-scalar');
                newline = this.buffer.length;
            }
        }
        // Trailing insufficiently indented tabs are invalid.
        // To catch that during parsing, we include them in the block scalar value.
        let index = newline + 1;
        unit = codeUnit(this.buffer, index);
        while(unit === 32) {
            index += 1;
            unit = codeUnit(this.buffer, index);
        }
        if(unit === 9) {
            while(unit === 9 || unit === 32 || unit === 13 || unit === 10) {
                index += 1;
                unit = codeUnit(this.buffer, index);
            }
            newline = index - 1;
        }
        else if(!this.blockScalarKeep) {
            while(true) {
                let lastIndex = newline - 1;
                let lastUnit = codeUnit(this.buffer, lastIndex);
                if(lastUnit === 13) {
                    lastIndex -= 1;
                    lastUnit = codeUnit(this.buffer, lastIndex);
                }
                const lastChar = lastIndex; // Drop the line if last char not more indented
                while(lastUnit === 32) {
                    lastIndex -= 1;
                    lastUnit = codeUnit(this.buffer, lastIndex);
                }
                if(lastUnit === 10 && lastIndex >= this.pos && lastIndex + 1 + indent > lastChar) newline = lastIndex;
                else break;
            }
        }
        this.tokens.push('\x1f');
        this.pushToIndex(newline + 1, true);
        return this.parseLineStart();
    }
    parsePlainScalar(): string {
        const inFlow = this.flowLevel > 0;
        let end = this.pos - 1;
        let index = this.pos - 1;
        let unit: number;
        while(true) {
            index++;
            unit = codeUnit(this.buffer, index);
            if(unit === -1) break;
            if(unit === 58) {
                const next = codeUnit(this.buffer, index + 1);
                if(isEmptyCode(next) || (inFlow && isFlowIndicatorCode(next))) break;
                end = index;
            }
            else if(isEmptyCode(unit)) {
                let next = codeUnit(this.buffer, index + 1);
                if(unit === 13) {
                    if(next === 10) {
                        index += 1;
                        unit = 10;
                        next = codeUnit(this.buffer, index + 1);
                    }
                    else end = index;
                }
                if(next === 35 || (inFlow && isFlowIndicatorCode(next))) break;
                if(unit === 10) {
                    const continuationStart = this.continueScalar(index + 1);
                    if(continuationStart === -1) break;
                    index = Math.max(index, continuationStart - 2); // to advance, but still account for ' #'
                }
            }
            else {
                if(inFlow && isFlowIndicatorCode(unit)) break;
                end = index;
            }
        }
        if(unit === -1 && !this.atEnd) return this.setNext('plain-scalar');
        this.tokens.push('\x1f');
        this.pushToIndex(end + 1, true);
        return inFlow ? 'flow' : 'doc';
    }
    parseDocument(): string {
        this.pushSpaces(true);
        const line = this.getLine();
        if(!this.lineAvailable) return this.setNext('doc');
        let count = this.pushIndicators();
        if(character(line, count) === '#' || character(line, count) === '') {
            if(character(line, count) === '#') this.pushCount(line.length - count);
            this.pushNewline();
            return this.parseLineStart();
        }
        switch(character(line, count)) {
            case '':
                this.pushNewline();
                return this.parseLineStart();
            case '{':
            case '[':
                this.pushCount(1);
                this.flowKey = false;
                this.flowLevel = 1;
                return 'flow';
            case '}':
            case ']':
                // this is an error
                this.pushCount(1);
                return 'doc';
            case '*':
                this.pushUntil(false);
                return 'doc';
            case '"':
            case "'":
                return this.parseQuotedScalar();
            case '|':
            case '>':
                count += this.parseBlockScalarHeader();
                count += this.pushSpaces(true);
                this.pushCount(line.length - count);
                this.pushNewline();
                return this.parseBlockScalar();
            default:
                return this.parsePlainScalar();
        }
    }
    parseFlowCollection(): string {
        let newline: number;
        let spaces: number;
        let indent = -1;
        do {
            newline = this.pushNewline();
            if(newline > 0) {
                spaces = this.pushSpaces(false);
                indent = spaces;
                this.indentValue = spaces;
            }
            else {
                spaces = 0;
            }
            spaces += this.pushSpaces(true);
        } while(newline + spaces > 0);
        const line = this.getLine();
        if(!this.lineAvailable) return this.setNext('flow');
        if(
            (indent !== -1 && indent < this.indentNext && character(line, 0) !== '#') ||
            (indent === 0 && (line.startsWith('---') || line.startsWith('...')) && isEmpty(character(line, 3)))
        ) {
            // Allowing for the terminal ] or } at the same (rather than greater)
            // indent level as the initial [ or { is technically invalid, but
            // failing here would be surprising to users.
            const atFlowEndMarker =
                indent === this.indentNext - 1 &&
                this.flowLevel === 1 &&
                (character(line, 0) === ']' || character(line, 0) === '}');
            if(!atFlowEndMarker) {
                // this is an error
                this.flowLevel = 0;
                this.tokens.push('\x18');
                return this.parseLineStart();
            }
        }
        let count = 0;
        while(character(line, count) === ',') {
            count += this.pushCount(1);
            count += this.pushSpaces(true);
            this.flowKey = false;
        }
        count += this.pushIndicators();
        switch(character(line, count)) {
            case '':
                return 'flow';
            case '#':
                this.pushCount(line.length - count);
                return 'flow';
            case '{':
            case '[':
                this.pushCount(1);
                this.flowKey = false;
                this.flowLevel += 1;
                return 'flow';
            case '}':
            case ']':
                this.pushCount(1);
                this.flowKey = true;
                this.flowLevel -= 1;
                return this.flowLevel !== 0 ? 'flow' : 'doc';
            case '*':
                this.pushUntil(false);
                return 'flow';
            case '"':
            case "'":
                this.flowKey = true;
                return this.parseQuotedScalar();
            case ':': {
                const next = this.charAt(1);
                if(this.flowKey || isEmpty(next) || next === ',') {
                    this.flowKey = false;
                    this.pushCount(1);
                    this.pushSpaces(true);
                    return 'flow';
                }
                this.flowKey = false;
                return this.parsePlainScalar();
            }
            default:
                this.flowKey = false;
                return this.parsePlainScalar();
        }
    }
    parseNext(next: string): string {
        switch(next) {
            case 'stream':
                return this.parseStream();
            case 'line-start':
                return this.parseLineStart();
            case 'block-start':
                return this.parseBlockStart();
            case 'doc':
                return this.parseDocument();
            case 'flow':
                return this.parseFlowCollection();
            case 'quoted-scalar':
                return this.parseQuotedScalar();
            case 'block-scalar':
                return this.parseBlockScalar();
            case 'plain-scalar':
                return this.parsePlainScalar();
        }
        return '';
    }
    lex(source: string, incomplete = false): string[] {
        this.tokens = [];
        if(source !== '') {
            this.buffer = this.buffer !== '' ? this.buffer + source : source;
            this.lineEndPos = -2;
        }
        this.atEnd = !incomplete;
        let next = this.next === '' ? 'stream' : this.next;
        while(next !== '') {
            if(!incomplete && !this.hasChars(1)) break;
            next = this.parseNext(next);
        }
        return this.tokens;
    }
}
