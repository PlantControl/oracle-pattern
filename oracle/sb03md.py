#!/usr/bin/env python3
"""Capture, replay, and explicitly audit the SB03MD migration."""
import argparse
import gzip
import json
import math
from pathlib import Path
import random
import sys
import subprocess
import tempfile

import run as common

ROOT = common.ROOT
BASELINE = ROOT / 'oracle/baselines/sb03md-v4'
POLICY = 'sb03md-float64-v2'
SOURCES = ['SB03MD','MB01RD','SB03MY','SB03MX','SB03MW','SB03MV','SB04PX']
GUARD = common.GUARD
EPS = common.EPS


def build_reference(work):
    return common.build_reference(work, routine='SB03MD', sources=SOURCES,
        driver='oracle/reference/sb03md.f90', capture_sources=['oracle/sb03md.py','oracle/run.py'])


def build_candidate(work):
    return common.build_candidate(work,name='sb03md',package='lyapunov',runner='oracle/sb03md.py')


def matrix(n, values):
    return [list(map(float, values[i*n:(i+1)*n])) for i in range(n)]


def flat(a):
    return common.strings([v for row in a for v in row])


def transpose(a):
    return list(map(list,zip(*a)))


def mul(a,b):
    return [[math.fsum(x*y for x,y in zip(row,col)) for col in zip(*b)] for row in a]


def identity(n):
    return [[float(i==j) for j in range(n)] for i in range(n)]


def case(name,n=3,dico='C',job='B',fact='F',trans='N',pad=2,rng=None,complex_blocks=True,scale=1):
    rng=rng or random.Random(19)
    a=[[0.0]*n for _ in range(n)]
    for i in range(n):
        for j in range(i+1,n):
            a[i][j]=rng.uniform(-.12,.12)
        a[i][i]=(-.8-.15*i) if dico.upper()=='C' else (.15+.035*i)
    if complex_blocks:
        for i in range(0,n-1,2):
            a[i+1][i+1]=a[i][i]
            a[i][i+1]=.2
            a[i+1][i]=-.3
    u=identity(n)
    for i in range(n-1):
        angle=rng.uniform(-2,2)
        co,si=math.cos(angle),math.sin(angle)
        for j in range(n):
            u[j][i],u[j][i+1]=co*u[j][i]-si*u[j][i+1],si*u[j][i]+co*u[j][i+1]
    if fact.upper()=='N':
        a=mul(mul(u,a),transpose(u))
        u=[[GUARD]*n for _ in range(n)]
    c=[[0.0]*n for _ in range(n)]
    for i in range(n):
        for j in range(i,n):
            c[i][j]=c[j][i]=scale*rng.uniform(-2,2)
    return dict(id=name,n=n,dico=dico,job=job,fact=fact,trans=trans,pad=pad,invalid=0,a=flat(a),u=flat(u),c=flat(c))


def corpus(seed,count):
    cases=[]
    for d in 'CD':
        for job in 'XSB':
            for fact in 'FN':
                for tr in 'NTC':
                    for n in (0,1,2,3,4,6):
                        cases.append(case(f'grid-{d}{job}{fact}{tr}-{n}',n,d,job,fact,tr))
    for d in 'cd':
        for job in 'xsb':
            cases.append(case(f'lower-{d}{job}',3,d,job,'f','t'))
    for bad in (1,2,4,5,7,9,11):
        for n in (0,3):
            c=case(f'invalid-{bad}-{n}',n);c['invalid']=bad;cases.append(c)
    for d in 'CD':
        for tr in 'NT':
            for value in (0.0,1.0,-1.0):
                c=case(f'scalar-{d}{tr}-{value}',1,d,'B','F',tr)
                c.update(a=[str(value)],u=['1.0'],c=['2.0']);cases.append(c)
            c=case(f'scaled-{d}{tr}',1,d,'B','F',tr)
            c.update(a=['1e-200' if d=='C' else '1.0'],u=['1.0'],c=['1e200']);cases.append(c)
        for magnitude in (1e-150,1e150):
            for fact in 'FN':
                c=case(f'rhs-{d}{fact}-{magnitude}',4,d,'B',fact,scale=magnitude);cases.append(c)
    for magnitude in (1e-200,1e200):
        c=case(f'extreme-schur-{magnitude}',2,'C','X','N')
        c['a']=common.strings([float(v)*magnitude for v in c['a']])
        c['c']=common.strings([float(v)*magnitude for v in c['c']]);cases.append(c)
    for d in 'CD':
        for tr in 'NT':
            c=case(f'opposite-eigenvalues-{d}{tr}',2,d,'B','F',tr)
            c.update(a=['1.0','0.0','0.0','-1.0'] if d=='C' else ['0.5','0.0','0.0','2.0'],u=flat(identity(2)),c=['1.0','2.0','2.0','3.0'])
            cases.append(c)
            for shape in ('real','complex'):
                c=case(f'large-rhs-{d}{tr}-{shape}',4,d,'B','F',tr,complex_blocks=shape=='complex',scale=1e290)
                cases.append(c)
            c=case(f'ignored-lower-{d}{tr}',4,d,'B','F',tr)
            for i in range(4):
                for j in range(i):c['c'][i*4+j]=str(12345+i*4+j)
                for j in range(i-1):c['a'][i*4+j]=str(-54321-i*4-j)
            cases.append(c)
    for d in 'CD':
        c=case(f'identity-anchor-{d}',2,d,'B','F')
        a=[[-.7,.2],[-.3,-.7]] if d=='C' else [[.2,.2],[-.3,.2]]
        rhs=mul(transpose(a),a) if d=='D' else [[a[i][j]+a[j][i] for j in range(2)] for i in range(2)]
        if d=='D':
            for i in range(2):rhs[i][i]-=1
        c.update(a=flat(a),u=flat(identity(2)),c=flat(rhs));cases.append(c)
    for d in 'CD':
        for tr in 'NT':
            for n in (2,4):
                cases.append(case(f'block-scaling-{d}{tr}-{n}',n,d,'B','F',tr,scale=1e300))
            special={
                'near':([-1e-12,1,-1,-1e-12] if d=='C' else [0,1-1e-12,-(1-1e-12),0]),
                'early-pivot':([0,1e-300,-1e-300,0] if d=='C' else [1,1e-100,-1e-100,1]),
                'final-pivot':[0,1,-1,0],
            }
            for label,a in special.items():
                c=case(f'{label}-{d}{tr}',2,d,'B','F',tr)
                c.update(a=common.strings(a),u=flat(identity(2)))
                if label=='final-pivot':c['c']=['0.0']*4
                cases.append(c)
    rng=random.Random(seed)
    for i in range(count):
        cases.append(case(f'random-{seed}-{i}',rng.randrange(1,10),rng.choice('CD'),rng.choice('XSB'),rng.choice('FN'),rng.choice('NTC'),rng.randrange(4),rng, bool(i%2),rng.choice([1,1e-100,1e100])))
    return cases


def validate_case(c):
    if set(c)!={'id','n','dico','job','fact','trans','pad','invalid','a','u','c'}:
        raise ValueError('invalid case fields')
    if not isinstance(c['id'],str) or not c['id']:
        raise ValueError('invalid case identity')
    for key,allowed in [('dico','CDcd'),('job','XSBxsb'),('fact','FNfn'),('trans','NTCntc')]:
        if not isinstance(c[key],str) or len(c[key])!=1 or c[key] not in allowed:
            raise ValueError('invalid mode')
    if any(type(c[k]) is not int or not 0<=c[k]<=256 for k in ('n','pad')):
        raise ValueError('invalid dimensions')
    if type(c['invalid']) is not int or c['invalid'] not in (0,1,2,4,5,7,9,11):
        raise ValueError('invalid probe')
    for key in ('a','u','c'):
        if not isinstance(c[key],list) or len(c[key])!=c['n']**2 or any(not isinstance(v,str) or not math.isfinite(float(v)) for v in c[key]):
            raise ValueError('invalid matrix '+key)


def reference_input(c):
    lines=[' '.join(str(c[k]) for k in ('dico','job','fact','trans','n','pad','invalid'))]
    n=c['n']
    if n:
        for key in ('a','u','c'):
            lines += [' '.join(c[key][i:i+n]) for i in range(0,n*n,n)]
    return '\n'.join(lines)+'\n'


def observe_reference(c,exe):
    tokens=common.invoke([str(exe)],data=reference_input(c)).split()
    n=c['n']
    if len(tokens)!=7+3*n*n+2*n:
        raise ValueError('reference observation length mismatch')
    info,handler,padding=map(int,tokens[:3])
    if padding not in (0,1):raise ValueError('invalid padding status')
    out=dict(info=info,handler=handler,padding_ok=bool(padding))
    for key,v in zip(('scale','sep','ferr'),tokens[3:6]):out[key]=repr(float(v))
    pos=6
    for key,length in [('a',n*n),('u',n*n),('c',n*n),('wr',n),('wi',n)]:
        out[key]=common.strings(list(map(float,tokens[pos:pos+length])));pos+=length
    out['true_sep']=repr(float(tokens[pos]))
    return out


def norm(a):
    return math.hypot(*(x for row in a for x in row))


def relative_error(a,b):
    scale=max(norm(a),norm(b))
    if scale==0:return 0.0
    return norm([[x/scale-y/scale for x,y in zip(r,s)] for r,s in zip(a,b)])


def schur_checks(c,o):
    n=c['n']
    a,u=matrix(n,o['a']),matrix(n,o['u'])
    if any(a[i][j]!=0 for i in range(n) for j in range(i-1)):
        return ['non-Hessenberg Schur form']
    failures=[]
    tol=512*EPS*max(1,n)
    for i in range(n-1):
        if a[i+1][i]!=0:
            if i+2<n and a[i+2][i+1]!=0:failures.append('overlapping Schur blocks')
            block_scale=max(abs(a[i][i]),abs(a[i+1][i+1]),abs(a[i][i+1]),abs(a[i+1][i]))
            if a[i][i+1]==0 or math.copysign(1,a[i][i+1])==math.copysign(1,a[i+1][i]):failures.append('invalid complex Schur block')
            if abs(a[i][i]/block_scale-a[i+1][i+1]/block_scale)>tol:failures.append('noncanonical complex Schur block')
    if relative_error(mul(transpose(u),u),identity(n))>tol:failures.append('Schur vectors not orthogonal')
    if relative_error(mul(mul(u,a),transpose(u)),matrix(n,c['a']))>tol:failures.append('Schur reconstruction failed')
    return failures


def residual(c,o):
    n=c['n']
    a=matrix(n,c['a']);x=matrix(n,o['c']);rhs=matrix(n,c['c'])
    for i in range(n):
        for j in range(i):rhs[i][j]=rhs[j][i]
    if c['fact'].upper()=='F':
        for i in range(n):
            for j in range(i-1):a[i][j]=0.0
        u=matrix(n,c['u']);a=mul(mul(u,a),transpose(u))
    if c['trans'].upper()!='N':a=transpose(a)
    # Normalize before multiplication so tiny/huge accepted inputs stay measurable.
    an,xn=norm(a),norm(x)
    if xn==0:return 0 if norm(rhs)==0 else math.inf
    aa=[[v/an if an else 0 for v in row] for row in a]
    xx=[[v/xn for v in row] for row in x]
    if c['dico'].upper()=='C':
        left,right=mul(transpose(aa),xx),mul(xx,aa)
        factor=(float(o['scale'])/an)/xn if an else math.inf
        diff=[[left[i][j]+right[i][j]-factor*rhs[i][j] for j in range(n)] for i in range(n)]
        return norm(diff)/(2+abs(factor)*norm(rhs))
    den=max(1,an)**2
    axa=mul(mul(transpose(aa),xx),aa)
    f=(float(o['scale'])/xn)/den
    diff=[[(an*an/den)*axa[i][j]-xx[i][j]/den-f*rhs[i][j] for j in range(n)] for i in range(n)]
    return norm(diff)/(an*an/den+1/den+abs(f)*norm(rhs))


def compare(c,ref,got):
    failures=[];n=c['n'];tol=4096*EPS*max(1,n);worst=0.0
    for label,o in [('reference',ref),('candidate',got)]:
        if type(o.get('info')) is not int or o.get('padding_ok') is not True:
            failures.append(label+': status/padding invalid');continue
        for key,length in [('a',n*n),('u',n*n),('c',n*n),('wr',n),('wi',n)]:
            if not isinstance(o.get(key),list) or len(o[key])!=length or any(not isinstance(v,str) or not math.isfinite(float(v)) for v in o[key]):
                failures.append(label+': malformed '+key)
        for key in ('scale','sep','ferr'):
            if not isinstance(o.get(key),str) or not math.isfinite(float(o[key])):failures.append(label+': malformed '+key)
    if failures:return failures,worst
    if ref.get('handler')!=-min(ref['info'],0):failures.append('reference error handler mismatch')
    if ref['info'] not in ((-c['invalid'],) if c['invalid'] else (0,n+1)):
        failures.append('unexpected reference INFO')
    if got['info']!=ref['info']:failures.append('INFO mismatch')
    for label,o in [('reference',ref),('candidate',got)]:
        for key in ('a','u','c'):
            unchanged=c['invalid'] or (key in ('a','u') and c['fact'].upper()=='F') or (key=='c' and c['job'].upper()=='S')
            if unchanged and not all(common.exact(x,y) for x,y in zip(o[key],c[key])):
                failures.append(label+': changed '+key)
    if c['invalid']:
        for key in ('scale','sep','ferr','wr','wi'):
            xs,ys=got[key],ref[key]
            if not isinstance(xs,list):xs,ys=[xs],[ys]
            if not all(common.exact(x,y) for x,y in zip(xs,ys)):failures.append('invalid-call output changed: '+key)
        return failures,worst
    for label,o in [('reference',ref),('candidate',got)]:
        for key in ('wr','wi'):
            if c['fact'].upper()=='F' and any(not common.exact(v,str(GUARD)) for v in o[key]):
                failures.append(label+': changed unused '+key)
        for key,unused in [('scale',n>0 and c['job'].upper()=='S'),('sep',c['job'].upper()=='X'),('ferr',c['job'].upper()!='B')]:
            if unused and not common.exact(o[key],str(GUARD)):failures.append(label+': changed unused '+key)
        if not n:
            for key,expected in [('scale',1),('sep',0 if c['job'].upper()!='X' else GUARD),('ferr',0 if c['job'].upper()=='B' else GUARD)]:
                if float(o[key])!=expected:failures.append(label+': wrong empty '+key)
    if not n:return failures,worst
    if c['fact'].upper()=='N':
        for label,o in [('reference',ref),('candidate',got)]:
            failures += [label+': '+f for f in schur_checks(c,o)]
            t=matrix(n,o['a'])
            ts=max(abs(v) for row in t for v in row) or 1.0
            t=[[v/ts for v in row] for row in t]
            spectrum=[];i=0
            while i<n:
                if i+1<n and t[i+1][i]!=0:
                    center=(t[i][i]+t[i+1][i+1])/2
                    disc=((t[i][i]-t[i+1][i+1])/2)**2+t[i][i+1]*t[i+1][i]
                    root=complex(disc)**.5
                    spectrum.extend([center+root,center-root]);i+=2
                else:spectrum.append(complex(t[i][i]));i+=1
            for wr,wi in zip(o['wr'],o['wi']):
                v=complex(float(wr)/ts,float(wi)/ts);j=min(range(len(spectrum)),key=lambda j:abs(spectrum[j]-v))
                expected=spectrum.pop(j)
                if abs(expected-v)>tol*max(abs(expected),abs(v)):failures.append(label+': eigenvalue disagrees with Schur form')
    if c['job'].upper()!='S':
        for label,o in [('reference',ref),('candidate',got)]:
            if not 0<float(o['scale'])<=1:failures.append(label+': invalid scale')
            x=matrix(n,o['c'])
            if relative_error(x,transpose(x))>tol:failures.append(label+': solution not symmetric')
            if o['info']==0 and residual(c,o)>tol:failures.append(label+': equation residual failed')
        ratio=relative_error(matrix(n,ref['c']),matrix(n,got['c']))/tol
        worst=max(worst,ratio)
        if ratio>1:failures.append(f'solution error/bound={ratio}')
    for key in ('scale','sep','ferr'):
        if key in ('sep','ferr') and c['fact'].upper()=='N' and c['job'].upper()!='X' and 'true_sep' in ref:
            continue
        x,y=float(ref[key]),float(got[key])
        ratio=abs(x-y)/(tol*max(abs(x),abs(y))) if x!=y else 0.0
        worst=max(worst,ratio)
        if ratio>1:failures.append(f'{key} mismatch {x} vs {y} (error/bound={ratio})')
    if c['fact'].upper()=='N' and c['job'].upper()!='X' and 'true_sep' in ref:
        sigma=float(ref['true_sep'])
        if not math.isfinite(sigma) or sigma<=0:failures.append('invalid independent separation')
        for label,o in [('reference',ref),('candidate',got)]:
            sep=float(o['sep'])
            if sep<=0:
                failures.append(label+': nonpositive separation');continue
            if not sigma/n*(1-tol)<=sep<=n*sigma*(1+tol):
                failures.append(label+': separation outside independent SVD bounds')
            if c['job'].upper()=='B':
                an=norm(matrix(n,o['a']))
                expected=EPS*an/sep if c['dico'].upper()=='C' else (EPS*an)*an/sep
                ratio=abs(float(o['ferr'])-expected)/(tol*abs(expected)) if expected else (0 if float(o['ferr'])==0 else math.inf)
                worst=max(worst,ratio)
                if ratio>1:failures.append(label+': forward error inconsistent with separation')
    return failures,worst


def trace(c):return dict(scenario=c['id'],caller='oracle/reference/sb03md.f90',callee='SB03MD',invocation=1)


def load_baseline(path):
    manifest=json.loads((path/'manifest.json').read_text())
    if manifest.get('schema')!=common.SCHEMA or manifest.get('policy') not in (POLICY,'sb03md-float64-v1') or manifest.get('routine')!='SB03MD' or manifest.get('kind')!='legacy-baseline':raise ValueError('unsupported baseline')
    if common.digest(path/'calls.jsonl.gz')!=manifest['calls_sha256']:raise ValueError('baseline checksum mismatch')
    with gzip.open(path/'calls.jsonl.gz','rt') as f:records=[json.loads(line) for line in f]
    if not records or len(records)!=manifest['count'] or len({r['case']['id'] for r in records})!=len(records):raise ValueError('empty/duplicate/wrong-count baseline')
    for r in records:
        validate_case(r['case'])
        if manifest['policy']==POLICY:
            value=r['expected'].get('true_sep')
            if not isinstance(value,str) or not math.isfinite(float(value)):
                raise ValueError('missing/invalid independent separation')
        if r['trace']!=trace(r['case']):raise ValueError('invalid caller trace')
        failures,_=compare(r['case'],r['expected'],r['expected'])
        if failures:raise ValueError(f"invalid recorded observation {r['case']['id']}: {failures}")
    return manifest,records


def capture(output,seeds,count,case_id=None):
    if output.exists():raise FileExistsError('baseline exists; choose a new version')
    cases={c['id']:c for seed in seeds for c in corpus(seed,count)}
    selected=common.select([dict(case=c) for c in cases.values()],case_id)
    records=[]
    with tempfile.TemporaryDirectory(prefix='sb03md-capture-') as work:
        exe,provenance=build_reference(Path(work))
        for r in selected:
            c=r['case'];validate_case(c);o=observe_reference(c,exe)
            failures,_=compare(c,o,o)
            if failures:raise ValueError(f"capture {c['id']}: {failures}")
            records.append(dict(case=c,expected=o,trace=trace(c)))
    output.mkdir(parents=True,exist_ok=False)
    payload=''.join(json.dumps(r,separators=(',',':'),allow_nan=False)+'\n' for r in records).encode()
    (output/'calls.jsonl.gz').write_bytes(gzip.compress(payload,mtime=0))
    manifest=dict(schema=common.SCHEMA,kind='legacy-baseline',routine='SB03MD',policy=POLICY,seeds=seeds,random_cases_per_seed=count,count=len(records),calls_sha256=common.digest(output/'calls.jsonl.gz'),reference=provenance)
    (output/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
    print(f'Captured {len(records)} calls: {output}')


def evaluate(baseline,report,case_id=None,*,audit=False):
    manifest,records=load_baseline(baseline);records=common.select(records,case_id)
    if report.resolve().is_relative_to(baseline.resolve()):raise ValueError('report must not overwrite baseline')
    results=[]
    with tempfile.TemporaryDirectory(prefix='sb03md-verify-') as work:
        exe,provenance=(build_reference if audit else build_candidate)(Path(work))
        for r in records:
            c=r['case']
            try:
                o=observe_reference(c,exe) if audit else json.loads(common.invoke([str(exe)],data=json.dumps(c)))
                failures,ratio=compare(c,r['expected'],o)
                if audit and o['handler']!=r['expected']['handler']:failures.append('error handler changed')
                if audit and 'true_sep' in r['expected']:
                    x,y=float(r['expected']['true_sep']),float(o['true_sep'])
                    if abs(x-y)>4096*EPS*max(1,c['n'])*max(abs(x),abs(y)):
                        failures.append('independent SVD separation changed')
                result=dict(case_id=c['id'],trace=r['trace'],actual=o,failures=failures,max_error_over_bound=ratio)
            except (RuntimeError,ValueError,KeyError,TypeError,OverflowError,OSError,subprocess.TimeoutExpired) as e:
                result=dict(case_id=c['id'],trace=r['trace'],failures=[str(e)],max_error_over_bound=None)
            if result['failures']:print(f"FAIL {c['id']}: {result['failures'][0]}",file=sys.stderr)
            results.append(result)
    failed=sum(bool(r['failures']) for r in results)
    summary=dict(calls=len(results),passed=len(results)-failed,failed=failed,max_error_over_bound=max((r['max_error_over_bound'] or 0 for r in results),default=0))
    report.parent.mkdir(parents=True,exist_ok=True)
    report.write_text(json.dumps(common.json_safe(dict(schema=common.SCHEMA,kind='audit' if audit else 'verification',policy=manifest['policy'],baseline=dict(manifest_sha256=common.digest(baseline/'manifest.json'),calls_sha256=manifest['calls_sha256'],seeds=manifest['seeds']),build=provenance,summary=summary,results=results)),indent=2,allow_nan=False)+'\n')
    print(json.dumps(summary));print(f'Report: {report}');return int(bool(failed))


def main():
    p=argparse.ArgumentParser(description=__doc__)
    sub=p.add_subparsers(dest='command',required=True)
    cap=sub.add_parser('capture');cap.add_argument('--output',type=Path,required=True);cap.add_argument('--seed',type=int,action='append');cap.add_argument('--random-cases',type=int,default=100);cap.add_argument('--case')
    for command in ('verify','audit'):
        s=sub.add_parser(command);s.add_argument('--baseline',type=Path,default=BASELINE);s.add_argument('--report',type=Path,default=ROOT/f'artifacts/sb03md-{command}.json');s.add_argument('--case')
    args=p.parse_args()
    if args.command=='capture':
        if args.random_cases<0:p.error('negative case count')
        capture(args.output,args.seed or [20260911,20260912],args.random_cases,args.case);return 0
    return evaluate(args.baseline,args.report,args.case,audit=args.command=='audit')


if __name__=='__main__':
    try:sys.exit(main())
    except (ValueError,RuntimeError,OSError,KeyError,TypeError,OverflowError,subprocess.TimeoutExpired) as e:
        print(f'sb03md oracle: {e}',file=sys.stderr);sys.exit(2)
