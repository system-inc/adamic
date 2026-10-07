import { readFileSync } from 'node:fs';
import { brotliCompressSync, constants, brotliDecompressSync } from 'node:zlib';
const input = readFileSync(process.argv[2]);
const compressed = brotliCompressSync(input, { params: { [constants.BROTLI_PARAM_QUALITY]: 11 } });
if (!brotliDecompressSync(compressed).equals(input)) throw new Error('Brotli round trip failed');
process.stdout.write(compressed);
