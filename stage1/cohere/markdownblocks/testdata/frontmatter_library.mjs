// Call the exact pinned library's Markdown parser, rather than duplicating its detector.
import fs from 'node:fs';
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
const markdown = require(process.argv[2] + '/plugins/markdown.js');
if(require(process.argv[2] + '/standalone.js').version !== '3.9.6') throw new Error('expected pinned Prettier 3.9.6');
const decode = text => text.replace(/\\([nrt\\])/g, (_, c) => c === 'n' ? '\n' : c === 'r' ? '\r' : c === 't' ? '\t' : c);
const encode = text => text.replace(/[\\\n\r\t]/g, c => c === '\\' ? '\\\\' : c === '\n' ? '\\n' : c === '\r' ? '\\r' : '\\t');
for(const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
    if(line === '') continue;
    const text = decode(line.slice(1));
    const root = markdown.parsers.markdown.parse(text, {});
    const frontMatter = root.children[0]?.type === 'frontMatter' ? root.children[0] : undefined;
    if(frontMatter === undefined) process.stdout.write('0\t' + encode(text) + '\n');
    else {
        const fields = ['1', encode(frontMatter.language), frontMatter.explicitLanguage === null ? '0' : '1' + encode(frontMatter.explicitLanguage),
            encode(frontMatter.value), encode(frontMatter.startDelimiter), encode(frontMatter.endDelimiter), encode(frontMatter.raw),
            encode(frontMatter.raw.replace(/[^\n]/g, ' ') + text.slice(frontMatter.raw.length))];
        process.stdout.write(fields.join('\t') + '\n');
    }
}
