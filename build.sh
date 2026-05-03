#!/bin/bash
set -e

cd "$(dirname "$0")"

build() {
    if [ -n "$1" ]; then
        VERSION="$1"
    else
        VERSION=$(date +%d%m%Y-%H%M%S)
    fi
    go build -ldflags "-X main.version=$VERSION" -o zerno ./cmd
    echo "Built: zerno ($VERSION)"
}

test() {
    echo "Running tests..."
    go test ./... "$@"
}

vet() {
    echo "Running vet..."
    go vet ./...
}

fmt() {
    echo "Formatting code..."
    go fmt ./...
}

coverage() {
    echo "Running tests with coverage..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report: coverage.html"
}

clean() {
    echo "Cleaning..."
    rm -f zerno
    find . -name "coverage.out" -o -name "coverage.html" -o -name "*.test" | xargs rm -f 2>/dev/null || true
}

run() {
    echo "Running zerno..."
    ./zerno "$@"
}

all() {
    fmt
    vet
    test
    build
}

help() {
    echo "Usage: ./build.sh <command>"
    echo ""
    echo "Commands:"
    echo "  build     - Build the binary"
    echo "  test      - Run tests"
    echo "  vet       - Run go vet"
    echo "  fmt       - Format code"
    echo "  coverage  - Run tests with coverage report"
    echo "  clean     - Clean build artifacts"
    echo "  run       - Run the built binary"
    echo "  all       - Format, vet, test, build"
    echo "  help      - Show this help"
}

case "${1:-help}" in
    build)    build ;;
    test)     test "${@:2}" ;;
    vet)      vet ;;
    fmt)      fmt ;;
    coverage) coverage ;;
    clean)    clean ;;
    run)      run "${@:2}" ;;
    all)      all ;;
    help)     help ;;
    *)        echo "Unknown command: $1"; help; exit 1 ;;
esac
