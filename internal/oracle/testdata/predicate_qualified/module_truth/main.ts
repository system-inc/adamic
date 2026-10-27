import * as ts from "./ts.ts";
type Value = number | string;
function use(value: Value): void { ts.assertNode(value, (x): x is number => typeof x === "number"); console.log("after"); }
use(7);
