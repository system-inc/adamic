function host(): void {
    function pick(a: number): number;
    function pick(a: number, f: (x: number) => string): string;
    function pick(a: number, f?: (x: number) => string): number | string {
        if (f) return f(a);
        return a;
    }
    function f(x: number): string { return "bad"; }
    const g: (a: number, f: (x: number) => string) => number = pick;
    const r = g(1, f);
    console.log(String(r + 1));
    console.log(r.toFixed(0));
}
host();
