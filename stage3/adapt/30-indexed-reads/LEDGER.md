- **2026-11-13, 30-indexed-reads, task 93w9rtz**: drop only the unread ambient
  require declaration from core's inserted header; keep process exactly unchanged.
  The Nov 8 amendment at docs/rulings.md:220 allowed that row only until its test
  disappeared; 99q dropped the test in 529705a1. Workers' checker survey found no
  reader across 79 roots. Adamic ruled Nov 13 09:08Z. Amend 32's embedded core
  measurement sources and hashes; add 30's check-mode guard and restoring mutant.
  internal/load/host_bindings.json's require row and all internal/ fixtures stay.
