#!/bin/bash -eu
# Checks for new upstream releases by comparing versions
# This script is used by the sync-upstream-releases workflow

set -o pipefail
export BASE="$(readlink -f "$(dirname "$0")/..")"

# Configuration
UPSTREAM_REPO="${UPSTREAM_REPO:-lrstanley/entrest}"
UPSTREAM_REMOTE="${UPSTREAM_REMOTE:-https://github.com/${UPSTREAM_REPO}.git}"

# Ensure upstream remote exists and fetch tags
setup_upstream() {
  if ! git remote get-url upstream >/dev/null 2>&1; then
    echo "Adding upstream remote: ${UPSTREAM_REMOTE}"
    git remote add upstream "${UPSTREAM_REMOTE}"
  fi

  echo "Fetching upstream tags..."
  git fetch upstream --tags

  echo "Fetching origin tags..."
  git fetch origin --tags
}

# Get current upstream version from upstream-sync/* tags
get_current_version() {
  # Get the latest upstream-sync/vX.Y.Z tag
  local latest_tag=$(git tag -l 'upstream-sync/v*' --sort=-v:refname | head -n1)

  if [ -z "$latest_tag" ]; then
    echo "Error: No upstream-sync/* tags found" >&2
    exit 1
  fi

  # Extract the version (everything after upstream-sync/)
  echo "$latest_tag" | sed 's|upstream-sync/||'
}

# Get latest upstream release tag (only stable releases, no pre-releases)
get_latest_version() {
  git ls-remote --tags --refs upstream | \
    grep -v '\^{}' | \
    awk '{print $2}' | \
    sed 's|refs/tags/||' | \
    grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | \
    sort -V | \
    tail -n1
}

# Determine update type based on version bump
get_update_type() {
  local old_ver="$1"
  local new_ver="$2"

  # Extract version numbers (remove 'v' prefix)
  old_ver="${old_ver#v}"
  new_ver="${new_ver#v}"

  # Split versions into major.minor.patch
  IFS='.' read -r old_major old_minor old_patch <<< "$old_ver"
  IFS='.' read -r new_major new_minor new_patch <<< "$new_ver"

  # Determine bump type
  if [ "$new_major" != "$old_major" ]; then
    echo "major"
  elif [ "$new_minor" != "$old_minor" ]; then
    echo "minor"
  else
    echo "patch"
  fi
}

# Main check logic
setup_upstream

CURRENT_VERSION=$(get_current_version)
LATEST_VERSION=$(get_latest_version)

echo "Current upstream version: ${CURRENT_VERSION}"
echo "Latest upstream version: ${LATEST_VERSION}"

# Determine update type if there's an update
if [ "${CURRENT_VERSION}" != "${LATEST_VERSION}" ]; then
  UPDATE_TYPE=$(get_update_type "${CURRENT_VERSION}" "${LATEST_VERSION}")
  echo "Update type: ${UPDATE_TYPE}"
fi

# Export for GitHub Actions
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "current_upstream=${CURRENT_VERSION}" >> "${GITHUB_OUTPUT}"
  echo "latest_upstream=${LATEST_VERSION}" >> "${GITHUB_OUTPUT}"

  if [ "${CURRENT_VERSION}" = "${LATEST_VERSION}" ]; then
    echo "needs_update=false" >> "${GITHUB_OUTPUT}"
  else
    echo "needs_update=true" >> "${GITHUB_OUTPUT}"
    echo "update_type=${UPDATE_TYPE}" >> "${GITHUB_OUTPUT}"
  fi
fi

if [ "${CURRENT_VERSION}" = "${LATEST_VERSION}" ]; then
  echo "Already up to date"
else
  echo "Update available: ${CURRENT_VERSION} -> ${LATEST_VERSION}"
fi

