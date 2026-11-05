let later: number | undefined = undefined;
bump();
console.log(String(later));
function bump(): void { if (later === undefined) { return; } later++; }
