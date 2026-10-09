Clean scope: timeout 120 go test -list . ./stage1/cohere/lint/rules/no-unsafe-negation/ > list.log 2>&1
Baseline and restored control: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/no-unsafe-negation/ -run .
Target coverage: same go test limits, -run '^TestCompileProfiles$' -coverpkg=github.com/system-inc/adamic/internal/load,github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native,github.com/system-inc/adamic/internal/javascript -coverprofile=target.cover
Rest coverage: same coverage flags with -run '^$' -coverprofile=rest.cover
Each run redirected stdout and stderr to its named .log. Source env.sh before Go commands. Mutation command, cache and wall time are in run.json; reproducible mutation script in mutate.py.
