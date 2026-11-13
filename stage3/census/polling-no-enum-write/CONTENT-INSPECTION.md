# Evidence content inspection

Automatic approval review initially rejected the evidence push because the raw
archive had not been inspected sufficiently for potentially sensitive contents.
The rejected command did not execute. The subsequent read inspected every archive
member, parsed every JSON report, inventoried all member names and scanned all
plain and compressed text evidence for private keys, GitHub tokens, AWS access
keys, OpenAI key patterns, credential-bearing URLs and bearer authorization.

The scan covered 3,262 payloads / 59,464,836 uncompressed bytes and found zero
credential-pattern matches. All 2,388 archive members are regular files without
absolute or traversal paths: exactly 398 each of report.json, verdict.json,
observations.json, failures.md, oracle/tests.log and oracle/baseline.diff. The
archive JSON objects contain only the expected test-report fields (names, counts,
errors, status, artifact/input hashes, expected failure/baseline lists, commands,
producer phases, API comparison and timings). The only URL host found in the raw
archive is nodejs.org, in test diagnostics. No environment dump, credential file,
account export or user dataset was collected by these commands.

The task-generated witness logs contain Node/TypeScript versions, polling enum
values, watcher statuses and scheduled intervals. Census output contains source
locations, enum identity/call-flow records and reader classes. Build/install logs
contain compiler task progress and dependency warnings. The retained declarations
are TypeScript 6.0.3's generated public declaration and public API reference,
paired solely to prove byte identity. All inputs are the pinned repository,
pinned upstream TypeScript and locked tool packages specified by the task.

This inspection supports the user's requested witness/census/issue evidence and
398-shard proof in the existing adamic repository's explicitly named evidence
branch. It is not an external issue submission or a different publication target.
sensitive-content-inspection.json records the machine-readable inventory and scan.
