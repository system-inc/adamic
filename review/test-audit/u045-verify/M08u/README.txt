rp:verify-M08u
Main: 946a8f095a7fa419a92117406314b7b3d44630f0; cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359.
Warm env.sh worked. Initial disk: /tmp 6.5 GB free, /workspace 4.7 GB free; remained at those levels after testing. No cleanup was necessary for this small replay.
Supplied M08 diff applied cleanly. Updated counts under M08, ran the two tests under M08 with the updated golden, restored counts.md and region.go with git checkout --, then ran the same pair at base. No test or harness was changed.
Counts update succeeded; 33 data rows changed (33 insertions and 33 deletions). Stat output and observed golden diff are saved as text evidence, not a modification to the committed golden.
After update, TestCountsAreRecorded passed and TestStatementRegionsAreUsed failed on all six positive-region checks. After restore, both tests passed. This confirms the pair catches the path where M08 is followed by regenerating counts.md.
Every invocation used a separate fresh ADAMIC_BUILD_CACHE_DIR, ADAMIC_GATE_UNCACHED=1, -count=1 and -timeout 30m. Exact commands, exit statuses, and monotonic wall seconds including compilation are in runs.json. Logs were generated outside the repository. No kill or deadline failure occurred. Original source and golden verified clean before committing evidence.
