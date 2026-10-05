# Third-party notices

Adamic is licensed `MIT OR Apache-2.0` ([LICENSE-MIT](LICENSE-MIT), [LICENSE-APACHE](LICENSE-APACHE)). Parts of it are
ported from other projects, each under its own license, reproduced here in full.

## In every program Adamic compiles

Adamic's runtime (`internal/native/runtime/`) is compiled into every native program Adamic builds, so a
binary Adamic makes carries what is ported into the runtime, and these notices travel with it.

### V8

- Source: https://github.com/v8/v8
- License: `BSD-3-Clause`
- In Adamic: number parsing (`runtime/parse.c`, after V8's src/numbers/conversions.cc) and exponentiation
  (`runtime/number.c`, after V8's math::pow), ported so Adamic's answers match Node's to the bit.

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

## In the compiler

The `adamic` compiler (stage 0) is built on cohere and the TypeScript compiler it carries (typescript-go,
Copyright (c) Microsoft Corporation, Apache License 2.0), and on the Go toolchain. Their notices are in
[cohere/NOTICE](cohere/NOTICE) and [cohere/THIRD_PARTY_NOTICES.md](cohere/THIRD_PARTY_NOTICES.md), and they travel with
any build of the compiler.
