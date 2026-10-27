type Value = number | string;
namespace N { export let f = (value: Value): value is number => true; }
function use(value: Value): void { if (N.f(value)) console.log("after"); }
use("text");
