M06 survives all three checks: all nine valid runs passed, and mutant timings were within 1.5 times both base runs. Logs are pushed to `test-audit/u045-M06i` under `review/test-audit/u045-M06i/`; the tree is restored. GNU time ran from an extracted Debian package, and JSON used the existing `Split_Setup`; initial evidence-file interference was corrected before repeating its sequence.

```json
{
  "mutant": "u045 M06",
  "base": "bc9edc55",
  "runs": [
    {"test":"TestDecodeASCIIUnit26","side":"base","exit":0,"seconds":43.162,"peak_rss_mb":1992.531,"signal":null,"line":null},
    {"test":"TestDecodeASCIIUnit26","side":"mutant","exit":0,"seconds":35.958,"peak_rss_mb":1992.926,"signal":null,"line":null},
    {"test":"TestDecodeASCIIUnit26","side":"base2","exit":0,"seconds":28.989,"peak_rss_mb":1991.379,"signal":null,"line":null},
    {"test":"TestPortMatchesGoCohere_123","side":"base","exit":0,"seconds":72.818,"peak_rss_mb":1997.957,"signal":null,"line":null},
    {"test":"TestPortMatchesGoCohere_123","side":"mutant","exit":0,"seconds":61.727,"peak_rss_mb":1995.590,"signal":null,"line":null},
    {"test":"TestPortMatchesGoCohere_123","side":"base2","exit":0,"seconds":70.425,"peak_rss_mb":1996.012,"signal":null,"line":null},
    {"test":"TestFileDriver_Setup","side":"base","exit":0,"seconds":69.028,"peak_rss_mb":1989.645,"signal":null,"line":null},
    {"test":"TestFileDriver_Setup","side":"mutant","exit":0,"seconds":69.406,"peak_rss_mb":1988.285,"signal":null,"line":null},
    {"test":"TestFileDriver_Setup","side":"base2","exit":0,"seconds":66.744,"peak_rss_mb":1989.332,"signal":null,"line":null}
  ],
  "caught_by": [],
  "survived": true
}
```
