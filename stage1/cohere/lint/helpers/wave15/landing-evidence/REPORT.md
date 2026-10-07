Rebased both retained CSS helpers onto origin/main e011f8f60899586d6373a5ccb07335ad82cfbf3c.

Uncached commands, with source /workspace/adamic-tools/env.sh:

- python3 stage1/cohere/lint/helpers/wave15/css_string/validate.py
- python3 stage1/cohere/lint/helpers/wave15/css_declaration/validate.py
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=20m
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v

All commands exited zero. CSS string compared 198,584 queries and 5,648,824 identical bytes; CSS declaration compared 8,836 queries and 619,181 identical bytes against unchanged Go cohere on source Node, emitted JavaScript and sanitized native. Both semantic mutants compiled, exited normally with empty stderr and were caught only by output comparison on all three backends. Original tests for all six Tailwind consumers passed in each run. Helper packages passed in 86.052s and 129.472s; filtered external-input oracle passed in 20.567s with six cache misses and no hits. The comment package also reran its semantic and explicit-refusal mutants.

This evidence does not certify the full repository gate, rule integration or unsupported JSX parser behavior. No new helper was claimed while either owned branch was awaiting landing.
