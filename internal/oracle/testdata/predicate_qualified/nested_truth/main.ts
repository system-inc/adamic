type Value = number | string;
namespace a { export namespace b { export function isX(value: Value): value is number { console.log("called"); return typeof value === "number"; } } }
function use(value: Value): void { if (a.b.isX(value)) console.log("after"); }
use(7);
