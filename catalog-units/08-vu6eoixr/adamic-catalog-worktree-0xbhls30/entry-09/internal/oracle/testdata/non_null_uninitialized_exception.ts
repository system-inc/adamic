function fail(): number { throw new Error('stop'); }
function run(): void {
    let value: number = undefined!;
    try { value = fail(); } catch { console.log('caught'); }
    console.log(`${value}`);
}
run();
