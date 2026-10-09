// Flips one bit in the middle of a file, in place: the smallest tamper there is.
//
//     node flip-byte.js <file>
//
// Every verification a release ships is shown refusing a file this has touched, because a check that has
// never once refused anything has not been shown to look at the bytes.
'use strict';

const fs = require('node:fs');

const file = process.argv[2];
const contents = fs.readFileSync(file);
if (contents.length === 0) {
    process.stderr.write(file + ' is empty, so there is no byte to flip and nothing a check could refuse\n');
    process.exit(1);
}
contents[contents.length >> 1] ^= 0x01;
fs.writeFileSync(file, contents);
