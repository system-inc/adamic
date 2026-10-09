// The external oracle runs only the pinned npm package, with no repository configuration.
import { createRequire } from 'node:module';
import { readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
const [directory, casesPath, answersPath] = process.argv.slice(2);
const require = createRequire(join(directory, 'package.json'));
const prettier = require('prettier');
if (prettier.version !== '3.9.6') {
    throw new Error(`expected Prettier 3.9.6, got ${prettier.version}`);
}
const answers = [];
const start = performance.now();
for (const item of JSON.parse(readFileSync(casesPath, 'utf8'))) {
    try {
        answers.push({ output: await prettier.format(item.text, { filepath: item.name }), error: '' });
    } catch (error) {
        answers.push({ output: '', error: error.message.split('\n')[0] });
    }
}
writeFileSync(answersPath, JSON.stringify(answers));
console.log(`${answers.length} texts in ${((performance.now() - start) / 1000).toFixed(3)}s`);
