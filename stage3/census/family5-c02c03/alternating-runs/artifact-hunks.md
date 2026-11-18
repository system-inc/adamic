# Built artifact differences

### built/local/_tsc.js.map
main bytes=2997654; C02 bytes=2997654; first difference at 0-based byte 64983.
```text
@@ bytes [64893:65143] around first difference @@
main: b'B,MAAtB;AACA,EAAAA,oDAAA,wBAAsB,MAAtB;AACA,EAAAA,oDAAA,cAAsB,MAAtB;AAXc,SAAAA;AAAA,GAAA;AAgwBX,IAAW,2BAAX,kBAAWC,8BAAX;AAEH,EAAAA,oDAAA,UAAO,KAAP;AACiB,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,YAAS,KAAT;AACA,EAAAA,oDAAA,UAAO,'
C02:  b'B,MAAtB;AACA,EAAAA,oDAAA,wBAAsB,MAAtB;AACA,EAAAA,oDAAA,cAAsB,MAAtB;AAXc,SAAAA;AAAA,GAAA;AAiwBX,IAAW,2BAAX,kBAAWC,8BAAX;AAEH,EAAAA,oDAAA,UAAO,KAAP;AACiB,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,YAAS,KAAT;AACA,EAAAA,oDAAA,UAAO,'
```

### built/local/compiler/tsconfig.tsbuildinfo
main bytes=53582; C02 bytes=53582; first difference at 0-based byte 25910.
```text
@@ bytes [25820:26070] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"7ae9a509881c3da4052e06388289a8e94c269576d38a8a808d9e08425c817239","signature":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat"'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"b994a750ae5eedd30b92326868848c91f7a5fd27179b56c3052b1cf04ec38c8e","signature":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat"'
```

### built/local/compiler/types.d.ts
main bytes=408078; C02 bytes=408209; first difference at 0-based byte 26127.
```diff
--- main/built/local/compiler/types.d.ts

+++ C02/built/local/compiler/types.d.ts

@@ -554,4 +554,5 @@

     /** @internal */ original?: Node | undefined;
     /** @internal */ emitNode?: EmitNode | undefined;
+    /** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;
 }
 export interface JSDocContainer extends Node {
```

### built/local/compiler/types.d.ts.map
main bytes=275854; C02 bytes=275989; first difference at 0-based byte 16640.
```text
@@ bytes [16550:16800] around first difference @@
main: b'C,QAAQ,CAAC,EAAE,IAAI,GAAG,SAAS,CAAC;IAC7C,gBAAgB,CAAC,QAAQ,CAAC,EAAE,QAAQ,GAAG,SAAS,CAAC;CAKpD;AAED,MAAM,WAAW,cAAe,SAAQ,IAAI;IACxC,oBAAoB,EAAE,SAAS,CAAC;IAChC,gBAAgB,CAAC,KAAK,CAAC,EAAE,UAAU,GAAG,SAAS,CAAC;CACnD;AAED,gBAAgB;AAChB,MAAM,WAAW,UAAW,SAAQ'
C02:  b'C,QAAQ,CAAC,EAAE,IAAI,GAAG,SAAS,CAAC;IAC7C,gBAAgB,CAAC,QAAQ,CAAC,EAAE,QAAQ,GAAG,SAAS,CAAC;IAKjD,gBAAgB,CAAC,6BAA6B,CAAC,EAAE,CAAC,CAAC,IAAI,EAAE,MAAM,EAAE,SAAS,EAAE,MAAM,EAAE,UAAU,CAAC,EAAE,IAAI,KAAK,MAAM,CAAC,GAAG,SAAS,CAAC;CACjI;AAED,MAAM,WAAW,cAAe'
```

### built/local/deprecatedCompat/tsconfig.tsbuildinfo
main bytes=47087; C02 bytes=47087; first difference at 0-based byte 36802.
```text
@@ bytes [36712:36962] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/harness/tsconfig.tsbuildinfo
main bytes=92987; C02 bytes=92987; first difference at 0-based byte 40974.
```text
@@ bytes [40884:41134] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/jsTyping/tsconfig.tsbuildinfo
main bytes=47708; C02 bytes=47708; first difference at 0-based byte 24549.
```text
@@ bytes [24459:24709] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/run.js.map
main bytes=6316408; C02 bytes=6316408; first difference at 0-based byte 90951.
```text
@@ bytes [90861:91111] around first difference @@
main: b'xD;AANO,MAAAH,gBAAS;AAAA,OAjTH;AAkUV,IAAM,eAAmD,eAAe;AAExE,IAAM,oBAAsD,eAAe;AAAA;AAAA;;;AC80N3E,SAAS,uBAAuB,GAAsC,YAAY,MAAc;AACnG,QAAM,OAAO,mBAAmB,EAAE,QAAQ;AAC1C,SAAO,YAAY,KAAK,YAAY,IAAI;AAC5C;AA5rOA,IAuCkBI,aAsuBAC,YAqEAC,gBAsDA,UAYAC,2BAeA,oBAivBA'
C02:  b'xD;AANO,MAAAH,gBAAS;AAAA,OAjTH;AAkUV,IAAM,eAAmD,eAAe;AAExE,IAAM,oBAAsD,eAAe;AAAA;AAAA;;;AC+0N3E,SAAS,uBAAuB,GAAsC,YAAY,MAAc;AACnG,QAAM,OAAO,mBAAmB,EAAE,QAAQ;AAC1C,SAAO,YAAY,KAAK,YAAY,IAAI;AAC5C;AA7rOA,IAuCkBI,aAsuBAC,YAqEAC,gBAsDA,UAYAC,2BAeA,oBAkvBA'
```

### built/local/server/tsconfig.tsbuildinfo
main bytes=80287; C02 bytes=80287; first difference at 0-based byte 36387.
```text
@@ bytes [36297:36547] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/services/tsconfig.tsbuildinfo
main bytes=62402; C02 bytes=62402; first difference at 0-based byte 22587.
```text
@@ bytes [22497:22747] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/testRunner/tsconfig.tsbuildinfo
main bytes=167608; C02 bytes=167608; first difference at 0-based byte 66662.
```text
@@ bytes [66572:66822] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/tsc/tsconfig.tsbuildinfo
main bytes=20556; C02 bytes=20556; first difference at 0-based byte 12086.
```text
@@ bytes [11996:12246] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/tsserver/tsconfig.tsbuildinfo
main bytes=80453; C02 bytes=80453; first difference at 0-based byte 36872.
```text
@@ bytes [36782:37032] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/typescript.internal.d.ts
main bytes=1546565; C02 bytes=1546700; first difference at 0-based byte 417061.
```diff
--- main/built/local/typescript.internal.d.ts

+++ C02/built/local/typescript.internal.d.ts

@@ -8767,4 +8767,5 @@

         /** @internal */ original?: Node | undefined;
         /** @internal */ emitNode?: EmitNode | undefined;
+        /** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;
     }
     interface Node {
```

### built/local/typescript.js.map
main bytes=4430249; C02 bytes=4430249; first difference at 0-based byte 93969.
```text
@@ bytes [93879:94129] around first difference @@
main: b'AA,YAAS,KAAT;AACA,EAAAA,wCAAA,WAAQ,KAAR;AACA,EAAAA,wCAAA,eAAY,KAAZ;AAJc,SAAAA;AAAA,GAAA;AAivBX,IAAW,2BAAX,kBAAWC,8BAAX;AAEH,EAAAA,oDAAA,UAAO,KAAP;AACiB,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,YAAS,KAAT;AACA,EAAAA,oDAAA,UAAO,'
C02:  b'AA,YAAS,KAAT;AACA,EAAAA,wCAAA,WAAQ,KAAR;AACA,EAAAA,wCAAA,eAAY,KAAZ;AAJc,SAAAA;AAAA,GAAA;AAkvBX,IAAW,2BAAX,kBAAWC,8BAAX;AAEH,EAAAA,oDAAA,UAAO,KAAP;AACiB,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,UAAO,KAAP;AACA,EAAAA,oDAAA,YAAS,KAAT;AACA,EAAAA,oDAAA,UAAO,'
```

### built/local/typescript/tsconfig.tsbuildinfo
main bytes=50910; C02 bytes=50910; first difference at 0-based byte 22307.
```text
@@ bytes [22217:22467] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/typingsInstaller/tsconfig.tsbuildinfo
main bytes=79935; C02 bytes=79935; first difference at 0-based byte 36718.
```text
@@ bytes [36628:36878] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```

### built/local/typingsInstallerCore/tsconfig.tsbuildinfo
main bytes=48370; C02 bytes=48370; first difference at 0-based byte 24934.
```text
@@ bytes [24844:25094] around first difference @@
main: b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"0bd3c78b4a2570b50f09d105c3c4247ff8d733082ba6668f71c2dcb3d7301a8c","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
C02:  b'c9b10c7f4e0943e28d761050632b84537266a813a6289ac0e201fccf2","impliedFormat":1},{"version":"71e2d3512563574fff3546200554da6fa24f8b1f9a73dacc5ab26bc948a6f3fa","impliedFormat":1},{"version":"48d54c870234330f3ece405c644c4c35383dda9c724a8efa547e51b4dfcaaa3'
```
