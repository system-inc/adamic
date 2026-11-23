- ff5eddz, adaptation 47 amendment: startTracing's three metadata literals build
  from explicit reads of the fresh local meta where upstream spread it after a
  field. They retain the same keys, in the same order, with the same values;
  the browser record's timeline category retains the position of meta's cat.
  The meta declaration and its single timestamp() call remain unchanged.
  The same JavaScript-changing amendment replaces typeCatalog.length = 0 with
  typeCatalog.splice(0), under rulings.md:306-307: the private dense array keeps
  its identity and becomes empty; the returned removed-elements array is unused.
  The newly visible JSON.stringify object references stop at 104:46 belongs
  to Compiler's #d9ep7de head 1; this amendment does not remove it.
  The complete combined 398-shard lane observes 0 startTracing runs; the
  positive control observes catalog length 3 before reset and 0 after, and
  meta own keys cat,ph,ts,pid,tid in that order.
