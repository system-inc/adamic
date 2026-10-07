// Parser-stage probe, not a whole-document formatter or a production driver.
import { panic, programArguments, readTextFile } from 'adamic';
import { parseFrontMatter } from '../frontmatter.ts';
import { decode, encode } from '../codec.ts';
const args = programArguments();
const source = readTextFile(args[0] ?? panic('usage: frontmatter_probe.ts <cases>'));
if(source.kind === 'Error') panic(source.message);
for(const line of source.text.split('\n')) {
    if(line === '') continue;
    const result = parseFrontMatter(decode(line.slice(1)));
    const frontMatter = result.frontMatter;
    if(frontMatter === undefined) console.log(`0\t${encode(result.content)}`);
    else {
        const fields: string[] = [
            '1',
            encode(frontMatter.language),
            frontMatter.explicitLanguage === undefined ? '0' : `1${encode(frontMatter.explicitLanguage)}`,
            encode(frontMatter.value),
            encode(frontMatter.startDelimiter),
            encode(frontMatter.endDelimiter),
            encode(frontMatter.raw),
            encode(result.content),
        ];
        console.log(fields.join('\t'));
    }
}
