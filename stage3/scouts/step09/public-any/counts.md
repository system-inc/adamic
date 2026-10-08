# Scout fixture counts

| Fixture | Node checks | Executable source mutants | Native claim |
|---|---:|---:|---|
| 01_primitive_config.a | stdout, stderr, exit | 1 | none |
| 02_unvalidated_list.a | stdout, stderr, exit | 1 | none |
| 03_timer_roundtrip.a | stdout, stderr, exit | 1 | none |

Total: three fixtures and three mutants. Each mutant exits zero with empty stderr,
then fails the unchanged Node stdout comparison. fixtures/status.json records
current-main compilation diagnostics. These probes are not internal/oracle
fixtures, so no native allocation/free table entry is added outside this territory.
