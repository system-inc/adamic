"""Observe evaluated buildcache recipes, using a compiler overlay on the candidate.

Only instrumentation is added. Product builds and key computation still use the
candidate's buildcache implementation. The checkout is never edited.
"""
import json
import os
import re
import signal
import subprocess


def bounded_run(command, seconds, **options):
    payload = options.pop("input", None)
    process = subprocess.Popen(command, start_new_session=True, stdin=subprocess.PIPE if payload is not None else None, **options)
    try:
        stdout, stderr = process.communicate(input=payload, timeout=seconds)
    except BaseException:
        os.killpg(process.pid, signal.SIGKILL)
        process.communicate()
        raise
    return subprocess.CompletedProcess(command, process.returncode, stdout, stderr)


def prepare_overlay(tree, directory):
    source = os.path.join(os.path.abspath(tree), 'internal/buildcache/buildcache.go')
    if not os.path.isfile(source):
        raise ValueError('candidate has no buildcache recipe API')
    with open(source) as handle:
        text = handle.read()
    signature = 'func get(inputs Inputs, build func(directory string) error) (string, string, error) {'
    if text.count(signature) != 1 or text.count('import (') != 1:
        raise ValueError('unsupported buildcache get API; cannot observe product inputs')
    text = text.replace('import (', 'import (\n gateRecipeJSON "encoding/json"', 1)
    text = text.replace(signature, signature.replace('func get(', 'func gateRecipeOriginalGet('), 1)
    text += '''
// Gate instrumentation: capture the actual recipe before its build can fail.
func get(inputs Inputs, build func(directory string) error) (string, string, error) {
    if path := os.Getenv("ADAMIC_GATE_PRODUCT_INPUTS"); path != "" {
        encoded, err := gateRecipeJSON.Marshal(inputs)
        if err != nil { return "", "", err }
        file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
        if err != nil { return "", "", err }
        _, err = file.Write(append(encoded, '\\n'))
        closed := file.Close()
        if err != nil { return "", "", err }
        if closed != nil { return "", "", closed }
        if os.Getenv("ADAMIC_GATE_PRODUCT_INPUTS_ONLY") == "1" {
            // Cached prerequisites may be read so a compound recipe can be
            // evaluated. A miss stops here without invoking any build callback.
            return gateRecipeOriginalGet(inputs, func(string) error {
                return fmt.Errorf("gate collected product inputs without building")
            })
        }
    }
    observedBuild := func(directory string) error {
        if path := os.Getenv("ADAMIC_GATE_PRODUCT_INPUTS"); path != "" && os.Getenv("ADAMIC_BUILD_CACHE") != "off" {
            file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
            if err != nil { return err }
            encoded, err := gateRecipeJSON.Marshal(map[string]string{"gate_product_outcome": "miss", "name": inputs.Name})
            if err == nil { _, err = file.Write(append(encoded, '\\n')) }
            file.Close()
            if err != nil { return err }
        }
        return build(directory)
    }
    return gateRecipeOriginalGet(inputs, observedBuild)
}
'''
    os.makedirs(directory, exist_ok=True)
    replacement = os.path.join(os.path.abspath(directory), 'buildcache.go')
    with open(replacement, 'w') as handle:
        handle.write(text)
    path = os.path.join(os.path.abspath(directory), 'overlay.json')
    with open(path, 'w') as handle:
        json.dump({'Replace': {source: replacement}}, handle)
    return path


def read_recipes(path):
    with open(path) as handle:
        rows = [json.loads(line) for line in handle if line.strip()]
        recipes = [row for row in rows if 'Name' in row and 'gate_product_outcome' not in row]
    if not recipes:
        raise ValueError('product unit did not report buildcache.Inputs')
    return recipes


def cold_miss(path):
    with open(path) as handle:
        return any(json.loads(line).get("gate_product_outcome") == "miss"
                   for line in handle if line.strip())
