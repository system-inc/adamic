// Both reused slices run separately. Together native regexp invalidates the
// cycle proof of the value parser's readonly child tree.
import { parse as parseSelector } from '../../selector/parser.ts';
import { parse as parseValue } from '../../values/values.ts';
console.log(parseSelector('.a').kind);
console.log(parseValue('red', true).kind);
