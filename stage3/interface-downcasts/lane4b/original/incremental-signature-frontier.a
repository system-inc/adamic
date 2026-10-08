import type { IncrementalMultiFileEmitBuildInfo } from 'original-tsc-builder';
interface Base { readonly fileNames: readonly string[]; }
function read(node: IncrementalMultiFileEmitBuildInfo): string { return typeof node.emitSignatures![0]; }
const concrete = {fileNames:['a'], root:[1], emitSignatures:[1], outSignature:'sig', fileInfos:['sig']};
const base: Base = concrete;
console.log(read(base as IncrementalMultiFileEmitBuildInfo));
