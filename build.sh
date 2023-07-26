set -euxo pipefail

mkdir -p "$(pwd)/cmd/gateway"
GOBIN=$(pwd)/cmd/gateway go install ./...
chmod +x "$(pwd)"/cmd/gateway/*