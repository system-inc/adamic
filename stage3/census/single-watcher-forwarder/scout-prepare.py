from pathlib import Path
import shutil,re,json,difflib
base=Path('/workspace/scratch/scout-adapted')
out=Path('/workspace/scratch/scout-514-1280/stage3/census/scout-514-1280')
s=(base/'src/compiler/sys.ts').read_text()
start=s.index('function createSingleWatcherPerName<')
end=s.index('\n/**',start)
helper=s[start:end]
old='''                (
                    // Cant infer types correctly so lets satisfy checker
                    (param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))
                ) as T,'''
assert old in helper
variants={}
# a: specialize the complete helper; each call selects its concrete helper.
parts=[]
for suffix,typ,p1,p2 in [('File','FileWatcherCallback','string','FileWatcherEventKind'),('Fs','FsWatchCallback','"rename" | "change"','string | undefined | null')]:
 h=helper.replace('createSingleWatcherPerName<T extends FileWatcherCallback | FsWatchCallback>',f'createSingle{suffix}WatcherPerName')
 h=re.sub(r'\bT\b',typ,h)
 h=h.replace(old.replace("as T", "as " + typ),f'                (param1: {p1}, param2: {p2}, param3?: Date) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3)),')
 parts.append(h)
a=s[:start]+'\n\n'.join(parts)+s[end:]
a=a.replace('return createSingleWatcherPerName(\n            pollingWatches,','return createSingleFileWatcherPerName(\n            pollingWatches,').replace('return createSingleWatcherPerName(\n            recursive ?','return createSingleFsWatcherPerName(\n            recursive ?')
variants['a']=a
# b: the common call domain of the union is string, never, optional Date.
b=helper.replace(old,'                (param1: "rename" | "change", param2: never, param3?: Date) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3)),')
variants['b']=s[:start]+b+s[end:]
# c: caller supplies a factory producing exactly T, consuming a snapshot getter.
c=helper.replace('    createWatcher: (callback: T) => FileWatcher,','    createWatcher: (callback: T) => FileWatcher,\n    createForwarder: (getCallbacks: () => T[] | undefined) => T,')
c=c.replace(old,'                createForwarder(() => cache.get(path)?.callbacks.slice()),')
c=s[:start]+c+s[end:]
for caller,typ,p1,p2 in [('cb => pollingWatchFileWorker(fileName, cb, pollingInterval, options),','FileWatcherCallback','string','FileWatcherEventKind'),('cb => fsWatchHandlingExistenceOnHost(fileOrDirectory, entryKind, cb, recursive, fallbackPollingInterval, fallbackOptions),','FsWatchCallback','"rename" | "change"','string | undefined | null')]:
 assert c.count(caller)==1
 c=c.replace(caller,caller+f'\n            (getCallbacks: () => {typ}[] | undefined): {typ} => (param1: {p1}, param2: {p2}, param3?: Date) => {{ const callbacks = getCallbacks(); if (callbacks) callbacks.forEach(cb => cb(param1, param2, param3)); }},')
variants['c']=c
for name,t in variants.items():
 root=Path('/workspace/scratch/scout-'+name)
 shutil.copytree(base/'src',root/'src')
 shutil.copytree(base/'lib',root/'lib')
 (root/'src/compiler/sys.ts').write_text(t)
 (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(s.splitlines(True),t.splitlines(True),fromfile='base/src/compiler/sys.ts',tofile=name+'/src/compiler/sys.ts')))
print('candidates prepared')
