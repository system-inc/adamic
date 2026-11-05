let later: number | undefined = undefined;
fill();
console.log(String(later));
function fill(): void { [later] = [4]; }
