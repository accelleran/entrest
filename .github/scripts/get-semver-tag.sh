#!/bin/bash -eu
# shellcheck disable=SC2155,SC2181
# Calculates the next semantic version tag with upstream metadata
# Outputs the version to stdout

set -o pipefail
export BASE="$(readlink -f "$(dirname "$0")/..")"
export BIN_DIR="${BASE}/.bin"

"${BASE}/scripts/install-svu.sh"

# Add local bin to PATH
export PATH="${BIN_DIR}:${PATH}"

# Pattern to filter only our fork's tags (with +upstream. metadata)
# This prevents svu from picking up plain tags from the upstream remote
TAG_PATTERN='v*+upstream.*'

# Get current version tag for informational purposes
CURRENT_TAG=$(svu current --tag.pattern="${TAG_PATTERN}" 2>/dev/null || echo "")
if [ -n "$CURRENT_TAG" ]; then
	echo "Current version tag: ${CURRENT_TAG}" >&2
else
	echo "Current version tag: (none - this will be the first)" >&2
fi

# Read upstream version from upstream-sync/* tags
UPSTREAM_SYNC_TAG=$(git tag -l 'upstream-sync/v*' --sort=-v:refname | head -n1)
if [ -z "$UPSTREAM_SYNC_TAG" ]; then
	echo "error: No upstream-sync/* tags found" >&2
	exit 1
fi

UPSTREAM_VERSION=$(echo "$UPSTREAM_SYNC_TAG" | sed 's|upstream-sync/v||')
if [ -z "$UPSTREAM_VERSION" ]; then
	echo "error: Could not extract upstream version from tag" >&2
	exit 1
fi

echo "Found latest merged upstream version: v${UPSTREAM_VERSION}" >&2

# Default to 'next' method if not specified
METHOD="${METHOD:-next}"
echo "Using method: ${METHOD}" >&2

# Determine base version based on method
case "$METHOD" in
	next)
		echo "" >&2
		echo "Analyzing commits for version bump:" >&2
		# Use --always to ensure we always bump at least patch, even with only chores
		BASE_TAG=$(svu next --always --verbose --tag.pattern="${TAG_PATTERN}")
		;;
	major | minor | patch)
		echo "" >&2
		echo "Analyzing commits for version bump:" >&2
		BASE_TAG=$(svu "$METHOD" --verbose --tag.pattern="${TAG_PATTERN}")
		;;
	alpha | rc)
		if [ -z "$CURRENT_TAG" ]; then
			echo "error: No current tag found for alpha/rc versioning" >&2
			exit 1
		fi

		read -r PR_TYPE REV <<< "$(sed -rn 's:^v?[0-9]+\.[0-9]+\.[0-9]+-(alpha|rc)\.([0-9]+)(\+.*)?$:\1 \2:p' <<< "$CURRENT_TAG")"

		if [ -z "$PR_TYPE" ] || [ -z "$REV" ] || [ "$PR_TYPE" != "$METHOD" ]; then
			# Starting a new alpha/rc sequence
			# Use 'next' to analyze commits and determine appropriate version bump
			REV=0
			PR_TYPE="$METHOD"
			METHOD="next"
		else
			# Incrementing existing alpha/rc sequence
			REV=$((REV + 1))
			METHOD="current"
		fi

		echo "" >&2
		echo "Analyzing commits for version bump:" >&2
		# Build alpha/rc tag
		BASE_TAG="$(svu "$METHOD" --verbose --tag.pattern="${TAG_PATTERN}")-${PR_TYPE}.${REV}"
		;;
	custom)
		BASE_TAG="$CUSTOM"
		echo "" >&2
		echo "Using custom version (no commit analysis)" >&2
		;;
	*)
		echo "error: unknown method" >&2
		exit 1
		;;
esac

# Add upstream metadata to base tag (unless already present)
if [[ "$BASE_TAG" =~ \+upstream\. ]]; then
	# Already has upstream metadata, use as-is
	NEW_TAG="$BASE_TAG"
else
	NEW_TAG="${BASE_TAG}+upstream.${UPSTREAM_VERSION}"
fi

# Show summary to user
echo "" >&2
echo "Next version tag: ${NEW_TAG}" >&2

# Compare current vs new base versions
CURRENT_BASE=$(echo "$CURRENT_TAG" | sed -E 's/\+upstream\..*$//')
if [ "$CURRENT_BASE" != "$BASE_TAG" ]; then
	echo "  Base version: ${CURRENT_BASE} → ${BASE_TAG}" >&2
else
	echo "  Base version: ${BASE_TAG} (unchanged)" >&2
fi
echo "  Upstream: v${UPSTREAM_VERSION}" >&2
echo "" >&2

# Output the version to stdout
echo "${NEW_TAG}"

# Export for GitHub Actions
if [ -n "${GITHUB_OUTPUT:-}" ]; then
	echo "tag=${NEW_TAG}" >> "${GITHUB_OUTPUT}"
	echo "base_version=${BASE_TAG}" >> "${GITHUB_OUTPUT}"
	echo "upstream_version=${UPSTREAM_VERSION}" >> "${GITHUB_OUTPUT}"
fi


