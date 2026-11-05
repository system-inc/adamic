function run(): void {
  let x: number | undefined = undefined;
  const set = (): void => { x = 4; };
  set();
  console.log(String(x));
}
run();
