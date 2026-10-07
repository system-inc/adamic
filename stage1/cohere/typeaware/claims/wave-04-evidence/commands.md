# Reproduction commands

Run from the repository root after sourcing `/workspace/adamic-tools/env.sh`.
The setup script selected that tool directory rather than `/opt/adamic-tools`.

```sh
bash cloud/setup.sh > /tmp/typeaware-wave-04-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc

gofmt -l cmd internal > /workspace/typeaware-wave-04/gofmt.log
# Also check the new checker files, oracle_wave_04.go and wave_04_test.go.
go vet ./... > /workspace/typeaware-wave-04/vet.log 2>&1
go test ./bridge/tsgo/... -count=1 -v -timeout=15m > /workspace/typeaware-wave-04/bridge-test.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout=10m > /workspace/typeaware-wave-04/node-oracle-test.log 2>&1
ADAMIC_WAVE04_ARTIFACTS=/workspace/typeaware-wave-04/final \
ADAMIC_WAVE04_REPOSITORY_MANIFEST=/workspace/typeaware-wave-04/repository.manifest \
ADAMIC_WAVE04_COMPILER_MANIFEST=/workspace/typeaware-wave-04/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typeaware-wave-04/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave04AgreementAndMutants$' -count=1 -v -timeout=30m > /workspace/typeaware-wave-04/final-test.log 2>&1
python3 bridge/tsgo/profile/volume_bench.py \
 /workspace/typeaware-wave-04/final/wave04 \
 /workspace/typeaware-wave-04/final/wave04-oracle \
 /workspace/typeaware-wave-04/bench \
 --corpus compiler /workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json /workspace/typeaware-wave-04/compiler.manifest \
 --corpus repository /workspace/adamic/tsconfig.json /workspace/typeaware-wave-04/repository.manifest \
 > /workspace/typeaware-wave-04/bench.log 2>&1
```

The compiler manifest is the 77 entries in `validation-coverage/compiler.manifest`,
resolved under TypeScript commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
The repository manifest is the frozen 287 entries in
`validation-coverage/repository.manifest`, resolved under this checkout.
`input-sha256.json` records their contents; new wave files are not added to the frozen corpus.
