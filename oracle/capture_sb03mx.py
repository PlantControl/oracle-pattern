#!/usr/bin/env python3
"""Explicitly capture SB03MX actions on the reference separation estimator probes."""
import argparse
import json
from pathlib import Path
import tempfile

import run
import sb03md


def capture(target):
    if target.exists():
        raise FileExistsError('probe recording exists; choose a new output')
    case = sb03md.case('grid-DSFN-6', 6, 'D', 'S', 'F', 'N')
    with tempfile.TemporaryDirectory(prefix='sb03mx-probes-') as directory:
        reference, provenance = run.build_reference(Path(directory), routine='SB03MX',
            sources=['SB03MX', 'SB03MV', 'SB04PX'], driver='oracle/reference/sb03mx.f90',
            capture_sources=['oracle/capture_sb03mx.py', 'oracle/sb03md.py'])
        probes = []
        for trans in 'NT':
            stdin = f'6 {trans}\n' + '\n'.join(' '.join(case['a'][i:i+6]) for i in range(0, 36, 6)) + '\n'
            lines = iter(run.invoke([str(reference)], data=stdin).splitlines())
            for line in lines:
                iteration, kase, op = line.split()
                before = list(map(float, next(lines).split()))
                info = int(next(lines))
                scale = float(next(lines))
                after = list(map(float, next(lines).split()))
                probes.append(dict(name=f'{trans}-{iteration}-{kase}', trans=op=='T', before=before, after=after, scale=scale, perturbed=info!=0))
        output = dict(provenance=provenance, a=list(map(float, case['a'])), probes=probes)
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(json.dumps(output, indent=2)+'\n')


if __name__ == '__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output',type=Path,required=True)
    capture(parser.parse_args().output)
