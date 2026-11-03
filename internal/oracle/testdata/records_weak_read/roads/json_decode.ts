// A road for a record into a weak read: the checked decode, decodeJson, into the record's own type,
// whose value is then handed to the weak type. (Decoded straight into Options it would be a fixed
// object, no road.)
import { decodeJson } from 'adamic';
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["overrideImportMode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly overrideImportMode?: 1 | 99 }
function mode(options: Options): string {
	return String(options.overrideImportMode);
}
const decoded = decodeJson<Table>('{"total":1,"overrideImportMode":"fast"}');
if (decoded.kind === "Ok") console.log(mode(decoded.value));
