import pathlib,json,subprocess,sys
r=pathlib.Path(sys.argv[1]).resolve(); o=pathlib.Path(sys.argv[2]).resolve()
subprocess.run(['python3',str(r/'stage3/census/latent/make_overlay.py'),str(r),str(o)],cwd=r,check=True)
m=json.loads((o/'overlay.json').read_text()); p=r/'stage3/census/append-constraint-scout'
for name in ['scan-only','signature']:
 target=o/(name+'.go');target.write_text((p/(name+'.go.txt')).read_text());m['Replace'][str(r/'internal/lower'/('append_scout_'+name.replace('-','_')+'.go'))]=str(target)
main=o/'stage3_census_latent_tool_main.go'
s=main.read_text().replace('return lower.Latent(context.Background(), program,','return lower.LatentScanOnly(context.Background(), program,')
s=s.replace('\tdiagnostics := program.LatentDiagnostics()', '\tif os.Getenv("SCOUT_SIGNATURE") == "1" { lower.ScoutSignature(context.Background(), program); return }\n\tdiagnostics := program.LatentDiagnostics()')
s=s.replace('\tprogram, err := load.LatentLoad(paths)', '\tif os.Getenv("SCOUT_SYS_ROOT") == "1" { paths = []string{filepath.Join(os.Args[1], "sys.ts")} }\n\tprogram, err := load.LatentLoad(paths)')
s=s.replace('\tif os.Getenv("LATENT_ASSERT_NO_OUTPUT")', '\tif os.Getenv("SCOUT_SYS_ROOT") == "1" { program.LatentReplayEntryReach() }\n\tif os.Getenv("LATENT_ASSERT_NO_OUTPUT")',1)
main.write_text(s);(o/'overlay.json').write_text(json.dumps(m,indent=2))
subprocess.run(['go','build','-buildvcs=false','-overlay='+str(o/'overlay.json'),'-o',sys.argv[3],'./stage3/census/latent/tool'],cwd=r,check=True)
