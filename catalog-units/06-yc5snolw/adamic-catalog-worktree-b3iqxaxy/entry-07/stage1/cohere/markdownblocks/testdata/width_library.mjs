// Original pinned width dependencies, composed as upstream get-string-width.js.
import fs from 'node:fs';
import {pathToFileURL} from 'node:url';
const root = process.argv[2];
const {default: emojiRegex} = await import(pathToFileURL(`${root}/node_modules/emoji-regex/index.js`));
const {_isFullWidth, _isWide} = await import(pathToFileURL(`${root}/node_modules/get-east-asian-width/index.js`));
const {narrowEmojiRegexp} = await import(pathToFileURL(`${root}/node_modules/narrow-emojis/index.js`));
const regex = emojiRegex();
function width(text) {
    if (!/[^\x20-\x7F]/.test(text)) return text.length;
    let result = 0;
    text = text.replace(regex, match => {result += narrowEmojiRegexp.test(match) ? 1 : 2; return '';});
    for (const char of text) {
        const code = char.codePointAt(0);
        if (code <= 0x1F || (code >= 0x7F && code <= 0x9F) || (code >= 0x300 && code <= 0x36F) || (code >= 0xFE00 && code <= 0xFE0F)) continue;
        result += _isFullWidth(code) || _isWide(code) ? 2 : 1;
    }
    return result;
}
const decode = text => text.replace(/\\(.)/g, (_, c) => c === 'n' ? '\n' : c === 'r' ? '\r' : c === 't' ? '\t' : c);
const output = [];
for (const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) if (line !== '') output.push(String(width(decode(line.slice(1)))));
process.stdout.write(output.join('\n') + '\n');
