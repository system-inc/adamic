interface Runner {
    run(): string;
}
let calls = 0;
class Lazy {
    get run(): () => string {
        calls++;
        return (): string => "ran";
    }
}
const runner: Runner = new Lazy();
console.log(runner.run ? "has run" : "no run");
console.log(String(calls));
