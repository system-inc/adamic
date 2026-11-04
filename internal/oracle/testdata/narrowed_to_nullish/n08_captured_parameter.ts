function run(x: number | undefined): void {
  const set = (): void => { x = 4; };
  x = undefined;
  set();
  console.log(String(x));
}
run(1);
