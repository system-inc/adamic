type Value = number | string;
function claim(value: Value): value is number { return true; }
namespace N { export let f = claim; }
function use(value: Value): void { if (N.f(value)) console.log("after"); }
use("text");
