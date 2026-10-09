# Consuming-rule ownership audit

Fetched all 1,418 origin references on 2026-10-08 and inspected every lint rule descriptor and every claim-containing path under stage1/cohere/lint. The only existing target rule directory was react-hooks/error-boundaries on origin/lint-helpers/rules-react at f1a5f571.

No other target had a rule directory or ownership claim. The ten distinct matching claim/evidence blobs were helper claims listing consumers, historical helper dependency snapshots, or selection records with empty claims. Those are not reservations of these rule ports.

Actual unique Go source/file/options capture counts:

| Rule | Count | Initial ownership |
| --- | ---: | --- |
| react/no-unused-state | 161 | free |
| react/default-props-match-prop-types | 111 | free |
| react-hooks/void-use-memo | 77 | free |
| react-hooks/error-boundaries | 69 | delivered here, f1a5f571 |
| react/no-access-state-in-setstate | 55 | free |
| react/no-set-state | 38 | free |

The counts come from unchanged Go tests with the existing RecordAssertedCase observer added by an overlay to rule_testing.go. Sources, file names and decoded options are deduplicated as in the lint harness. Refresh ownership before each free rule is started.
