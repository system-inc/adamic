import { readFileSync } from 'node:fs';
import { brotliDecompressSync } from 'node:zlib';
const input = readFileSync(process.argv[2]);
const compressed = readFileSync(0);
if (!brotliDecompressSync(compressed).equals(input)) throw new Error('Brotli round trip failed');
