#!/bin/bash

set -e

# 定义颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}$1${NC}"
}

# 打印错误信息
error() {
    echo -e "${RED}$1${NC}"
}

# 运行 go fmt
info "Running go fmt..."
if ! go fmt ./...; then
    error "go fmt failed"
    exit 1
fi

# 运行 go vet
info "Running go vet..."
if ! go vet ./...; then
    error "go vet failed"
    exit 1
fi

# 运行 go lint (需要安装 golangci-lint)
info "Running golangci-lint..."
if ! golangci-lint run; then
    error "golangci-lint failed"
    exit 1
fi

# 运行 go test
trap "rm coverage.out" EXIT
COVERAGE_THRESHOLD=70

go test -p 1 -coverprofile=coverage.out ./...

COVERAGE=$(go tool cover -func=coverage.out | awk '/total:/ {print substr($3, 1, length($3)-1)}')

info "Total test coverage: ${COVERAGE}%"

if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
  error "Coverage below threshold: ${COVERAGE_THRESHOLD}%"
  exit 1
else
  info "Coverage meets the threshold: ${COVERAGE_THRESHOLD}%"
fi