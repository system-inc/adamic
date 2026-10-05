# Third-party notices

Adamic is licensed `MIT OR Apache-2.0` ([LICENSE-MIT](LICENSE-MIT), [LICENSE-APACHE](LICENSE-APACHE)). Parts of it are
ported from other projects, each under its own license, reproduced here in full.

## In every program Adamic compiles

Adamic's runtime (`internal/native/runtime/`) is compiled into every native program Adamic builds, so a
binary Adamic makes carries what is ported into the runtime, and these notices travel with it.

### V8

- Source: https://github.com/v8/v8
- License: `BSD-3-Clause`
- In Adamic, ported so Adamic's answers match Node's to the bit:
  - number parsing (`runtime/parse.c`, after src/numbers/conversions.cc);
  - exponentiation (`runtime/number.c`, after math::pow);
  - V8's changes to fdlibm's Math functions (`runtime/ieee754.c`, after src/base/ieee754.cc; fdlibm's own
    notice is below);
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

## In the compiler

The `adamic` compiler (stage 0) is built on cohere and the TypeScript compiler it carries (typescript-go,
Copyright (c) Microsoft Corporation, Apache License 2.0), and on the Go toolchain. Their notices are in
[cohere/NOTICE](cohere/NOTICE) and [cohere/THIRD_PARTY_NOTICES.md](cohere/THIRD_PARTY_NOTICES.md), and they travel with
any build of the compiler.
