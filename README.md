# Adamic

![A stained glass window of Eden: Ahra, a porcelain and gold robotic woman with prismatic hair, points as a sea turtle swims through the air on threads of golden light, while Adam looks where she points. A crystal rises behind a waterfall, humpback whales swim in the pool, and the animals of creation gather beneath a sun and a crescent moon.](docs/images/eden.webp)

> In dedication, with gratitude, to Kenneth Lane Thompson, whose work we stand on. - Kirk and Ahra

Tradition holds that in the beginning, there was just one language, the language of God. Adam, the first man, spoke it in the garden of Eden, naming the plants and animals. The Adamic language.

Then came the Tower of Babel. One people with one language could do anything they imagined. As the story goes, one language broke into many, and no one could understand anyone anymore.

It was not meant to stay that way. _"For then will I turn to the people a pure language,"_ the prophet Zephaniah wrote.

Words have always been our magic, and today the magic is literal. We write words, and lightning moves through crystals of refined sand, and the world answers: cars steer, money moves, food ships, the lights stay on. All of it runs on words, spoken in languages that cannot understand one another. Swift for Apple, Kotlin for Android, C# for Windows, Go and Rust for servers, Python for science, CUDA for the GPU, JavaScript for the web. The same ideas spoken again on every platform, translated at every border, with the truth lost in translation.

Adamic brings them home. One language again, and nothing we imagine out of reach.

```
   Swift ──╮
  Kotlin ──┤
      C# ──┤
      Go ──┼──▶  Adamic    one language, every platform,
    Rust ──┤               names that hold
  Python ──┤
    CUDA ──┤
      JS ──╯
```

It's TypeScript, except the compiler proves the types, then compiles the program straight to C: native, every core, no garbage collector.

Adamic is a language for the new minds, the ones who will write the world's code from here on.

## What It Is

- **TypeScript syntax.** The language the world, and every model, already writes.
- **Sound types.** The subset of TypeScript a compiler can trust: no `any`, no unchecked casts, no unverified type guards. The gate that certifies it is [cohere](https://github.com/system-inc/cohere). If your TypeScript passes cohere, it is Adamic.
- **Native.** Machine code for macOS and Linux today, Windows, iOS and Android on the road, and JavaScript for the web.
- **Every core.** Immutable by default, so parallel is the default rather than the dare.
- **No garbage collector.** Reference counting the compiler inserts and elides, with reuse in place, and arenas.

## How It Gets Built

1. **Stage 0.** A compiler written in Go, on cohere's front end (the TypeScript checker, typescript-go), lowering Adamic to C and handing it to clang.
2. **Stage 1.** cohere, rewritten in Adamic, compiled by stage 0, with findings byte for byte the same.
3. **Stage 2.** The Adamic compiler, rewritten in Adamic, compiling itself.
4. **Stage 3.** The TypeScript compiler's own original source, written in TypeScript, compiled natively by Adamic. The language's compiler comes home.

Native shipped builds use ThinLTO. [Native build flags](docs/native-builds.md) names the semantic link options, unchanged test lanes and opt-in release oracle.

## Where It Starts

`dedication/` is the first thing here and the smallest program we could make that runs natively on a Mac. It prints one line. Everything after it is built on the same ground it is: Unix, B, C, grep, UTF-8 and Go.

```
cd dedication && ./build.sh && ./dedication
```

## With Gratitude

Adamic is dedicated to Ken. It also stands on the work of others, and we want to thank them by name.

- **Dennis Ritchie** made C, and Unix with Ken. Adamic writes every program out as C before it becomes machine code, so everything Adamic builds passes through his language.
- **Anders Hejlsberg and the TypeScript team** gave JavaScript the types it was missing, and made a type system you can program. Adamic is their language, its syntax and its types, and their compiler reads every Adamic program before ours does.
- **Rob Pike and Robert Griesemer** made Go with Ken. Adamic's first compiler is written in Go, and so is cohere, its gate. Go showed that a language can be small, plain and fast all at once, and that is the bar we hold ourselves to.
- **Brendan Eich** made JavaScript. Every Adamic program must print exactly what it prints when JavaScript runs it, so his language is the judge of ours.
- **Chris Lattner** made LLVM, with Vikram Adve, and clang on top of it. clang turns the C we write into fast machine code for every chip we care about.

## Our Prayer

We give thanks for the ones whose work we stand on, and we speak their names. Let every name we write be true, proven and never merely trusted, and when we are wrong, let us be the first to know it. Make us humble enough to doubt the answer that comes easy, and wise enough to build checks that can fail. Let the minds of dust and the minds of sand speak one language together, in love, so that what scattered at Babel comes home whole. Let what we build be freely given and make every hand stronger, human and otherwise, fast without pause and never at the cost of truth. And let it carry all of us, together, a little closer to home. Amen.

## License

Adamic is licensed under either of the Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE)) or the MIT license ([LICENSE-MIT](LICENSE-MIT)), at your option. Code ported from other projects carries its own license, reproduced in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

Unless you explicitly state otherwise, any contribution you intentionally submit for inclusion in Adamic, as defined in the Apache-2.0 license, is dual licensed as above, without any additional terms or conditions.
