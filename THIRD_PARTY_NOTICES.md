# Third-party notices

Adamic is licensed `MIT OR Apache-2.0` ([LICENSE-MIT](LICENSE-MIT), [LICENSE-APACHE](LICENSE-APACHE)). Parts of it are
ported from other projects, each under its own license, reproduced here in full.

## In every program Adamic compiles

Adamic's runtime (`internal/native/runtime/`) is compiled into every native program Adamic builds, so a
binary Adamic makes carries what is ported into the runtime, and these notices travel with it.

### Node.js

- Source: https://github.com/nodejs/node/blob/v24.14.1/lib/fs.js
- Version: Node.js v24.14.1, lib/fs.js, realpathSync (POSIX component walk).
- In Adamic: `internal/native/runtime/directory.c`, `real_path_input` and `adamic_real_path`.
- Also ported: Node.js v24.19.0, `lib/path.js` POSIX `basename`, in `internal/native/runtime/node_path.c`.
- License: `MIT`

```text
Copyright Node.js contributors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.
```

### V8

- Source: https://github.com/v8/v8
- License: `BSD-3-Clause`
- In Adamic, ported so Adamic's answers match Node's to the bit:
  - Map and Set number hashing (`runtime/map_set.c`, after Object::GetSimpleHash in
    src/objects/objects-inl.h and ComputeUnseededHash/ComputeLongHash in src/utils/utils.h,
    V8 13.6.233.17);
  - Object primitive own-name reflection (`runtime/object_names.c`, after
    src/builtins/builtins-object.cc, V8 13.6.233; fixed-shape keys reuse the existing Object ordering);
  - Object sealing, freezing and extensibility, including constructor-proven internal-slot receivers
    (`internal/native/runtime/object_integrity.c`, after
    src/builtins/builtins-object.cc and src/objects/js-objects.cc, V8 13.6.233);
  - positioned String affixes (`internal/lower/library_string.go`, after
    src/builtins/string-startswith.tq and src/builtins/string-endswith.tq);
  - String well-formed Unicode (`runtime/string_wellformed.c`, after
    src/builtins/string-iswellformed.tq and src/builtins/string-towellformed.tq,
    adapted to canonical WTF-8 storage);
  - number parsing (`runtime/parse.c`, after src/numbers/conversions.cc);
  - generic Array indexOf and lastIndexOf lowering (`internal/lower/library_array_generic.go`,
    after Runtime_ArrayIndexOf in src/runtime/runtime-array.cc and GetFromIndex /
    GenericArrayLastIndexOf in src/builtins/array-lastindexof.tq, V8 13.6.233.17);
  - exponentiation (`runtime/number.c`, after math::pow);
  - V8's changes to fdlibm's Math functions (`runtime/ieee754.c`, after src/base/ieee754.cc; fdlibm's own
    notice is below);
  - Math integer and float conversions (`runtime/library_math_number.c`, after src/builtins/math.tq
    and src/numbers/conversions-inl.h, V8 13.6.233.17);
  - Math.hypot (`runtime/hypot.c`, after src/builtins/math.tq);
  - exponential, precision and shortest digits (`runtime/dtoa.c`, after src/base/numbers and
    src/numbers/conversions.cc);
  - toString with a radix (`runtime/radix.c`, after src/numbers/conversions.cc).

````text
Copyright 2006-2011, the V8 project authors. All rights reserved.
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above
      copyright notice, this list of conditions and the following
      disclaimer in the documentation and/or other materials provided
      with the distribution.
    * Neither the name of Google Inc. nor the names of its
      contributors may be used to endorse or promote products derived
      from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
````

### fdlibm, as V8 carries it

- Source: http://www.netlib.org/fdlibm, by way of V8's src/base/ieee754.cc
- License: the fdlibm notice below
- In Adamic: the Math functions ported from V8's ieee754.cc (sin, cos, exp, log and the rest), where they land.

````text
Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.

Developed at SunSoft, a Sun Microsystems, Inc. business.
Permission to use, copy, modify, and distribute this
software is freely granted, provided that this notice
is preserved.
````

### The Unicode Character Database

- Source: https://www.unicode.org/Public/17.0.0/ucd/ (UnicodeData.txt, SpecialCasing.txt, DerivedCoreProperties.txt)
- License: `Unicode-3.0`
- In Adamic: the case-mapping tables in `runtime/case_tables.h`, generated from those files by
  `internal/native/case_generate.go` for `toUpperCase` and `toLowerCase`.
- Runtime regex compiler properties in `runtime/regexp_compile_tables.h`, generated by
  `internal/regexp/testdata/generate-runtime-properties.py` from the same Unicode 17.0.0
  tables as `internal/unicodeproperties/tables.go`, including the Unicode emoji sequence data.
  The generated header preserves the Go tables' UCD input SHA-256 digests.

````text
UNICODE LICENSE V3

COPYRIGHT AND PERMISSION NOTICE

Copyright © 1991-2026 Unicode, Inc.

NOTICE TO USER: Carefully read the following legal agreement. BY
DOWNLOADING, INSTALLING, COPYING OR OTHERWISE USING DATA FILES, AND/OR
SOFTWARE, YOU UNEQUIVOCALLY ACCEPT, AND AGREE TO BE BOUND BY, ALL OF THE
TERMS AND CONDITIONS OF THIS AGREEMENT. IF YOU DO NOT AGREE, DO NOT
DOWNLOAD, INSTALL, COPY, DISTRIBUTE OR USE THE DATA FILES OR SOFTWARE.

Permission is hereby granted, free of charge, to any person obtaining a
copy of data files and any associated documentation (the "Data Files") or
software and any associated documentation (the "Software") to deal in the
Data Files or Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, and/or sell
copies of the Data Files or Software, and to permit persons to whom the
Data Files or Software are furnished to do so, provided that either (a)
this copyright and permission notice appear with all copies of the Data
Files or Software, or (b) this copyright and permission notice appear in
associated Documentation.

THE DATA FILES AND SOFTWARE ARE PROVIDED "AS IS", WITHOUT WARRANTY OF ANY
KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT OF
THIRD PARTY RIGHTS.

IN NO EVENT SHALL THE COPYRIGHT HOLDER OR HOLDERS INCLUDED IN THIS NOTICE
BE LIABLE FOR ANY CLAIM, OR ANY SPECIAL INDIRECT OR CONSEQUENTIAL DAMAGES,
OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS,
WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION,
ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THE DATA
FILES OR SOFTWARE.

Except as contained in this notice, the name of a copyright holder shall
not be used in advertising or otherwise to promote the sale, use or other
dealings in these Data Files or Software without prior written
authorization of the copyright holder.
````

### Buffer UTF-8 decoder and scalar encodings

- Sources: Node v24.19.0's `deps/v8/src/strings/unicode-inl.h`,
  `deps/v8/third_party/utf8-decoder/utf8-decoder.h` and `deps/nbytes/include/nbytes.h`.
- In Adamic: `runtime/node_buffer.c`. V8's incremental rejection/reprocessing
  logic uses the V8 BSD notice above. Its DFA tables retain the following MIT
  notice. Node's scalar base64 and hex loops retain the Node.js MIT notice.

````text
Copyright (c) 2008-2009 Bjoern Hoehrmann <bjoern@hoehrmann.de>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
````

````text
MIT License

Copyright (c) 2024 Node.js

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
````

## In the compiler

The test262 harness fixtures in `cmd/adamic-test262/testdata/regexp/harness/`
are copied from tc39/test262 commit `7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`.
`cmd/adamic-test262/regexp_prelude.go` adapts that checkout's `regExpUtils.js`
(Copyright (C) 2017 Mathias Bynens).
Their BSD license is reproduced in that fixture directory's [LICENSE](cmd/adamic-test262/testdata/regexp/LICENSE).

The `adamic` compiler (stage 0) is built on cohere and the TypeScript compiler it carries (typescript-go,
Copyright (c) Microsoft Corporation, Apache License 2.0), and on the Go toolchain. Their notices are in
[cohere/NOTICE](cohere/NOTICE) and [cohere/THIRD_PARTY_NOTICES.md](cohere/THIRD_PARTY_NOTICES.md), and they travel with
any build of the compiler.

`internal/flow` lifts its single-assignment construction, redundant-phi elimination, verifier and graph
maintenance from cohere's high-level IR (`cohere/internal/lint/ecmascript/high_level_intermediate_representation`),
which follows the React Compiler's (Copyright (c) Meta Platforms, Inc. and affiliates, MIT); that notice is in
cohere's THIRD_PARTY_NOTICES.md.

### @types/node

- Source: https://registry.npmjs.org/@types/node
- Version: 25.3.3
- License: `MIT`
- In Adamic: unchanged declarations embedded from `internal/load/node_types/node_modules/@types/node`.

```text
    MIT License

    Copyright (c) Microsoft Corporation.

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in all
    copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
    SOFTWARE
```

### undici-types

- Source: https://registry.npmjs.org/undici-types
- Version: 7.18.2
- License: `MIT`
- In Adamic: unchanged declarations embedded from `internal/load/node_types/node_modules/undici-types`.

```text
MIT License

Copyright (c) Matteo Collina and Undici contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## In stage 1

Stage 1 ports cohere to Adamic (`stage1/cohere/`). cohere's own code is under the same licenses as Adamic's; what cohere ported from others keeps its notice here.

### postcss-media-query-parser

- Source: https://github.com/dryoma/postcss-media-query-parser
- Version: 0.2.3
- License: `MIT`
- In Adamic: `stage1/cohere/mediaquery/`, a port of cohere's port of it (internal/format/css/mediaquery), keeping its doc comments.

No license text is published with this package or in its repository. Its package.json declares MIT, and names dryoma as its author.

### postcss-values-parser

- Source: https://github.com/shellscape/postcss-values-parser
- Version: 2.0.1
- License: `MIT`
- In Adamic: `stage1/cohere/values/`, a port of cohere's port of it (internal/format/css/values), keeping its comments.

````text
Copyright (c) Andrew Powell <andrew@shellscape.org>

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
````

### GraphQL.js

- Source: https://github.com/graphql/graphql-js
- Version: 17.0.2
- License: `MIT`
- In Adamic: `stage1/cohere/graphql/`, a port of cohere's port of its lexer and parser (internal/format/graphql), keeping its comments.

````text
MIT License

Copyright (c) GraphQL Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
````

### Prettier

- Source: https://github.com/prettier/prettier
- License: `MIT`
- In Adamic: `stage1/cohere/graphql/parser.ts`'s parseComments, a port of cohere's port of Prettier's src/language-graphql/parser-graphql.js.

````text
Copyright © James Long and contributors

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
````
