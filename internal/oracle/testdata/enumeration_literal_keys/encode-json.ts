import "./support/iterator.ts";
import { encodeJson } from "adamic";
interface Row { one: number }
console.log(encodeJson<Row>({ one: 1 }));
