#!/bin/bash
set -e

cd "$(dirname "$0")"

build() {
    VERSION=$(date +%d%m%Y-%H%M%S)
    go build -ldflags "-X main.version=$VERSION" -o zerno ./cmd
    echo "Built: zerno ($VERSION)"
}

install_bin() {
    sudo install -m 755 zerno /usr/local/bin/zerno.new
    sudo mv /usr/local/bin/zerno.new /usr/local/bin/zerno
    echo "Installed: /usr/local/bin/zerno"
}

echo "Formatting code..."
go fmt ./...

echo "Running vet..."
go vet ./...

echo "Running tests..."
go test ./...

echo "Checking python syntax..."
for f in assets/files/*.py assets/conf/*.py; do
    python3 -c "import ast,sys; ast.parse(open(sys.argv[1]).read())" "$f" || exit 1
done

build

if [ "$1" != "--no-install" ]; then
    install_bin
fi
