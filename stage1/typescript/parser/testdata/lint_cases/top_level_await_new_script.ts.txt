await new X();
await new X;
f(await new X());
f(await new X);
const called = await new X();
const uncalled = await new X;
(await new X());
(await new X);
const sum = 1 + await new X();
const choice = true ? await new X() : await new X;
await 1;
await 1n;
await "value";
await true;
new X();
async function inside() { await new X(); await new X; }
function ordinary() { await new X(); await new X; }
