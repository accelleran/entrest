#!/bin/bash -eu
# shellcheck disable=SC2155

set -o pipefail
export BASE="$(readlink -f "$(dirname "$0")/..")"
export BIN_DIR="${BASE}/.bin"

mkdir -p "${BIN_DIR}"

if [ -f "${BIN_DIR}/svu" ]; then
	exit 0
fi

# renovate: datasource=github-releases depName=caarlos0/svu
SVU_VERSION="3.2.4"

echo "installing svu ${SVU_VERSION} to ${BIN_DIR}"
curl -sSL "https://github.com/caarlos0/svu/releases/download/v${SVU_VERSION}/svu_${SVU_VERSION}_linux_amd64.tar.gz" \
	| tar -C "${BIN_DIR}" -xzvf- svu

echo "svu installed successfully"

