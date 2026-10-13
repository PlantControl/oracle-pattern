#!/usr/bin/env python3
"""Cross-feed independently computed Schur factors to isolate estimator differences."""
import json
from pathlib import Path
import tempfile
import sb03md
import run


def main():
    c = sb03md.case('grid-DSNT-6', 6, 'D', 'S', 'N', 'T')
    with tempfile.TemporaryDirectory(prefix='sb03md-factors-') as directory:
        reference, legacy_build = sb03md.build_reference(Path(directory) / 'legacy')
        candidate, go_build = sb03md.build_candidate(Path(directory) / 'go')
        legacy = sb03md.observe_reference(c, reference)
        go = json.loads(run.invoke([str(candidate)], data=json.dumps(c)))
        observations = []
        for name, factors in [('fortran', legacy), ('go', go)]:
            supplied = dict(c, fact='F', a=factors['a'], u=factors['u'])
            f = sb03md.observe_reference(supplied, reference)
            g = json.loads(run.invoke([str(candidate)], data=json.dumps(supplied)))
            failures,_=sb03md.compare(supplied,f,g)
            if failures:raise ValueError(f'{name} factors: {failures}')
            observations.append(dict(factor_source=name, fortran=f, go=g))
        out = dict(case=c, computed_fortran=legacy, computed_go=go,
            cross_fed=observations, legacy_build=legacy_build, go_build=go_build)
        target = run.ROOT / 'artifacts/sb03md-factors.json'
        target.parent.mkdir(parents=True,exist_ok=True)
        target.write_text(json.dumps(out, indent=2)+'\n')
        for observation in observations:
            print(observation['factor_source'], observation['fortran']['sep'], observation['go']['sep'])
        print(target)


if __name__ == '__main__':
    main()
