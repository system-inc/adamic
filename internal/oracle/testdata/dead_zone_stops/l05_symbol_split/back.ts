import { other } from '../l05_symbol_split.ts';
export const splitter = { [Symbol.split](text: string): string[] { return [other(), text]; } };
