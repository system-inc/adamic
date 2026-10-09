// A proving program: 2.2.3's namespace() does not consume this bar.
import { createRequire } from 'node:module';
import { join } from 'node:path';
import { writeSync } from 'node:fs';
const [directory, text] = process.argv.slice(2);
const require = createRequire(join(directory, 'package.json'));
const Processor = require('postcss-selector-parser/dist/processor.js');
writeSync(1, 'entering parser\n');
new Processor(() => {}).process(text);
writeSync(1, 'returned\n');
