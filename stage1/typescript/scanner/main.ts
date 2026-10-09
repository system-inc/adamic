// File driver. --manifest runs many independent files in one process for the corpus and timings.
import { panic, programArguments, readTextFile, utf8Length } from 'adamic';
import { Scanner } from './scanner.ts';

function written(text: string): string {
    let result = '';
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        result +=
            code >= 32 && code <= 126 && code !== 92
                ? text.slice(index, index + 1)
                : `\\u${code.toString(16).padStart(4, '0')}`;
    }
    return result;
}
function run(path: string, mode: string, countOnly: boolean): number {
    const read = readTextFile(path);
    if(read.kind === 'Error') {
        panic(read.message);
    }
    // A single linear mapping makes every reported position a Go byte offset, including diagnostics.
    const offsets: number[] = [0];
    // Count-only runs do not report byte positions, so they need no offset table.
    if(!countOnly) {
        let bytes = 0;
        for(let index = 0; index < read.text.length; index++) {
            const code = read.text.codePointAt(index) ?? 0;
            if(code > 0xffff) {
                offsets.push(bytes, bytes + 4);
                index++;
                bytes += 4;
            }
            else {
                bytes += code < 128 ? 1 : code < 2048 ? 2 : 3;
                offsets.push(bytes);
            }
        }
        if(bytes !== utf8Length(read.text)) {
            panic('source byte mapping differs');
        }
    }
    const scanner = new Scanner(read.text);
    let count = 0;
    let errorIndex = 0;
    for(;;) {
        if(mode === 'jsx') {
            scanner.scanJsx();
        }
        else {
            scanner.scan();
        }
        if(mode === 'regex') {
            scanner.rescanSlash();
        }
        if(mode === 'greater') {
            scanner.rescanGreater();
        }
        if(mode === 'template' && scanner.kind === 'CloseBraceToken') {
            scanner.rescanTemplate();
        }
        count++;
        if(!countOnly) {
            while(errorIndex < scanner.errors.length) {
                const error = scanner.errors[errorIndex] ?? panic('missing diagnostic');
                const start = offsets[error.start] ?? panic('diagnostic start outside source');
                const end = offsets[error.start + error.length] ?? panic('diagnostic end outside source');
                console.log(`error ${error.code} ${start} ${end - start}`);
                errorIndex++;
            }
            const hasValue =
                scanner.kind === 'Identifier' ||
                scanner.kind === 'PrivateIdentifier' ||
                scanner.kind.endsWith('Keyword') ||
                scanner.kind === 'StringLiteral' ||
                scanner.kind === 'NumericLiteral' ||
                scanner.kind === 'BigIntLiteral' ||
                scanner.kind === 'NoSubstitutionTemplateLiteral' ||
                scanner.kind === 'TemplateHead' ||
                scanner.kind === 'TemplateMiddle' ||
                scanner.kind === 'TemplateTail' ||
                scanner.kind === 'RegularExpressionLiteral' ||
                scanner.kind === 'JsxText' ||
                scanner.kind === 'JsxTextAllWhiteSpaces';
            console.log(
                `${scanner.kind} ${offsets[scanner.start] ?? panic('token start outside source')} ${offsets[scanner.pos] ?? panic('token end outside source')} ${scanner.flags}${hasValue ? `\t${written(scanner.value)}` : ''}`,
            );
        }
        if(scanner.kind === 'EndOfFile') {
            break;
        }
    }
    return count;
}
const args = programArguments();
const first =
    args[0] ?? panic('usage: main.ts <source file> [scan|regex|greater|template|jsx] or --manifest <file> [--count]');
if(first === '--manifest') {
    const manifest = readTextFile(args[1] ?? panic('missing manifest'));
    if(manifest.kind === 'Error') {
        panic(manifest.message);
    }
    const countOnly = args[2] === '--count';
    let count = 0;
    let caseNumber = 0;
    for(const line of manifest.text.split('\n')) {
        if(line === '') {
            continue;
        }
        const fields = line.split('\t');
        if(!countOnly) {
            console.log(`case ${caseNumber}`);
        }
        count += run(fields[1] ?? panic('missing source'), fields[0] ?? 'scan', countOnly);
        caseNumber++;
    }
    if(countOnly) {
        console.log(`${count}`);
    }
}
else {
    run(first, args[1] ?? 'scan', false);
}
