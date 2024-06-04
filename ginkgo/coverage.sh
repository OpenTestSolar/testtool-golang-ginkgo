#!/bin/bash
trap "rm coverage.out" EXIT
COVERAGE_THRESHOLD=70

go test -p 1 -coverprofile=coverage.out ./...

COVERAGE=$(go tool cover -func=coverage.out | awk '/total:/ {print substr($3, 1, length($3)-1)}')

echo "Total test coverage: ${COVERAGE}%"

if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
  echo "Coverage below threshold: ${COVERAGE_THRESHOLD}%"
  exit 1
else
  echo "Coverage meets the threshold: ${COVERAGE_THRESHOLD}%"
fi