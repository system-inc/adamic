#!/usr/bin/env bash
set -euo pipefail
out=$1
pin=050880ce59e30b356b686bd3144efe24f875ebc8
scratch=$(mktemp -d)
trap 'rm -rf "${scratch:?}"' EXIT
curl --fail --location --silent --show-error --max-time 35 "https://codeload.github.com/microsoft/TypeScript/tar.gz/$pin" -o "$scratch/source.tar.gz"
node - "$scratch/source.tar.gz" <<'JS'
const fs=require('node:fs'),crypto=require('node:crypto');
const sha=crypto.createHash('sha256').update(fs.readFileSync(process.argv[2])).digest('hex');
if(sha!=='43d388f7fc6511ae105f23f4779e1c8df96f6dfa003025b70249816089609f5b')throw Error('TypeScript archive hash changed');
JS
tar -xzf "$scratch/source.tar.gz" --strip-components=1 -C "$out" "TypeScript-$pin/src/compiler" "TypeScript-$pin/scripts/processDiagnosticMessages.mjs"
(cd "$out" && node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json)
node - "$out" "$(dirname "$0")/parse_manifest.json" <<'JS'
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const manifest=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));
if(manifest.typescript_commit!=='050880ce59e30b356b686bd3144efe24f875ebc8'||manifest.files.length!==77)throw Error('wrong parse corpus');
for(const file of manifest.files){const relative=file.path.replace(/^typescript\//,'');const data=fs.readFileSync(path.join(process.argv[2],relative));if(crypto.createHash('sha256').update(data).digest('hex')!==file.sha256)throw Error('changed input '+relative);}
JS
