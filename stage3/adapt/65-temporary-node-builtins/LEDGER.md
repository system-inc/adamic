- #nqp4p7g, amendments to 47 and 65: under Adamic's #00qn5az ruling,
  tracing.ts loads fs through a qualified module-scope node:fs namespace
  import, following adaptation 93's builtin-import precedent. This is
  JavaScript-changing. Removing the sole lazy require leaves an empty try;
  its original "tracing requires having fs" rethrow is unreachable, so the
  complete empty loader guard is removed. 47 pins the new finished owner
  and validates all qualified bindings; 65 stops adding ambient require
  and keeps the readonly process.pid declaration. Evidence belongs on
  typescript/tracing-fs-import-evidence.
