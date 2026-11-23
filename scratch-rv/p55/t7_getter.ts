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
if (runner.run) {
    console.log(runner.run());
}
console.log(String(calls));
