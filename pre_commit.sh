#!/bin/bash
set -e

# Run standard tests (which runs go tests by default if any)
go test ./... -short

# We should also run python tests in the venv
source chronos_v3/venv/bin/activate
PYTHONPATH=. pytest chronos_v3/tests/
