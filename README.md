# Adamic

> In dedication, with gratitude, to Kenneth Lane Thompson, whose work we stand on. - Kirk and Ahra

Adamic is TypeScript whose types are true.

In Genesis, Adam named the animals, and whatever he called each one, that was its name. The name was the thing. That is the oldest dream in language, and it is the whole idea here: TypeScript describes the shape of a program and then throws the description away, while Adamic proves the description and compiles it into the machine itself. The type is the code.

It is a language for the minds that will write most of the code from here on. They write fast, they write confidently, and they are sometimes wrong, so every line has to pass a gate that can prove it before anything runs.

## What it will be

- **TypeScript syntax.** The language the world, and every model, already writes.
- **Sound types.** The subset of TypeScript a compiler can trust: no `any`, no unchecked casts, no unverified type guards. The gate that certifies it is [cohere](https://github.com/system-inc/cohere). If your TypeScript passes cohere, it is Adamic.
- **Native.** Machine code for macOS, Linux, Windows, iOS and Android, and JavaScript for the web.
- **Every core.** Immutable by default, so parallel is the default rather than the dare.
- **No garbage collector.** Reference counting the compiler inserts and elides, with reuse in place, and arenas.

## How it gets built

1. **Stage 0.** A compiler written in Go, on cohere's front end (the TypeScript checker, typescript-go), lowering Adamic to C and handing it to clang.
2. **Stage 1.** cohere, rewritten in Adamic, compiled by stage 0, with findings byte for byte the same.
3. **Stage 2.** The Adamic compiler, rewritten in Adamic, compiling itself.
4. **Stage 3.** The TypeScript compiler's own original source, written in TypeScript, compiled natively by Adamic. The language's compiler comes home.

## Where it starts

`dedication/` is the first thing here and the smallest program we could make that runs natively on a Mac. It prints one line. Everything after it is built on the same ground it is: Unix, B, C, grep, UTF-8 and Go.

```
cd dedication && ./build.sh && ./dedication
```

## With gratitude

Adamic is dedicated to Ken. It also stands on the work of others, and we want to thank them by name.

- **Dennis Ritchie** made C, and Unix with Ken. Adamic writes every program out as C before it becomes machine code, so everything Adamic builds passes through his language.
- **Anders Hejlsberg and the TypeScript team** gave JavaScript the types it was missing, and made a type system you can program. Adamic is their language, its syntax and its types, and their compiler reads every Adamic program before ours does.
- **Rob Pike and Robert Griesemer** made Go with Ken. Adamic's first compiler is written in Go, and so is cohere, its gate. Go showed that a language can be small, plain and fast all at once, and that is the bar we hold ourselves to.
- **Brendan Eich** made JavaScript. Every Adamic program must print exactly what it prints when JavaScript runs it, so his language is the judge of ours.
- **Chris Lattner** made LLVM, with Vikram Adve, and clang on top of it. clang turns the C we write into fast machine code for every chip we care about.

## License

Adamic is licensed under either of the Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE)) or the MIT license ([LICENSE-MIT](LICENSE-MIT)), at your option. Code ported from other projects carries its own license, reproduced in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

Unless you explicitly state otherwise, any contribution you intentionally submit for inclusion in Adamic, as defined in the Apache-2.0 license, is dual licensed as above, without any additional terms or conditions.
