function make(): string { return String(new Late() instanceof Late); }
class Late { static first = make(); }
console.log(String(Late.first));
