Starting commands: git fetch origin && git checkout -b test-defend/stage1-cohere-yaml-scalars_match_go_shards origin/main; git fetch origin test-audit/stage1-cohere-yaml-scalars_match_go_shards:refs/remotes/origin/test-audit/stage1-cohere-yaml-scalars_match_go_shards.

Warm source /workspace/adamic-tools/env.sh; npm ci --prefix stage3/api. Installed yaml2.9.0 prettier3.9.6 yaml-unist-parser3.2.1 in /tmp/yaml-defense/library. Lockfile retained.

Coverage: ADAMIC_YAML_LIBRARY=/tmp/yaml-defense/library timeout120 go test -json -count=1 -timeout90s -coverpkg=./internal/native,./internal/lower -coverprofile=cost.cover ./stage1/cohere/yaml/ -run ^TestSpeedCostProbes$; corresponding scalar profile runs ^TestScalarsMatchGo(Union|_[0-9]{3})$. Exact logs retain paths and observations.

D2 matrix: ADAMIC_YAML_LIBRARY=/tmp/yaml-defense/library ADAMIC_BUILD_CACHE_DIR=/tmp/yaml-defense/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestSpeedCostProbes|TestScalarsMatchGo(Union|_[0-9]{3})|TestSchemaMatchesGo|TestWidthsMatchGo)$'
D1 same matrix without Schema/Widths and with cache/D1; cooked. D1 narrowed rerun only ^TestScalarsMatchGo(Union|_[0-9]{3})$.

Cost witness: timeout20 /tmp/yaml-defense/D1-string-probe /tmp/yaml-defense/units.txt slice1 (Bash time); control same argument/mode/rounds on native binary built from restored source. Both have checksum529984000. No tests/harness/oracles modified.
