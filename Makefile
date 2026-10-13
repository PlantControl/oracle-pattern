.PHONY: verify audit verify-offline

verify:
	go test ./...
	go vet ./...
	python3 -m unittest discover -s oracle -p 'test_oracle.py' -v
	python3 -m unittest discover -s oracle -p 'test_sb03md.py' -v
	python3 oracle/run.py verify
	python3 oracle/sb03md.py verify

audit:
	python3 -m unittest discover -s oracle -p 'test_capture.py' -v
	python3 -m unittest discover -s oracle -p 'test_sb03md_capture.py' -v
	python3 oracle/run.py audit
	python3 oracle/sb03md.py audit
	python3 oracle/audit_sb03md_factors.py

verify-offline:
	python3 oracle/check_offline.py
