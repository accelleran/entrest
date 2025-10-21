# Versioning Strategy

This fork of [lrstanley/entrest](https://github.com/lrstanley/entrest) uses a custom versioning scheme to track both our fork's changes and the upstream version we're based on.

## Version Format

```
vX.Y.Z+upstream.A.B.C
```

Where:
- `X.Y.Z` = Our fork's semantic version
- `A.B.C` = The upstream version we're based on

### Examples

- `v1.0.0+upstream.1.0.2` - Initial fork from upstream v1.0.2
- `v1.1.0+upstream.1.0.2` - Added new feature, still based on upstream v1.0.2
- `v1.2.0+upstream.1.1.0` - Merged upstream v1.1.0 release

## Semantic Versioning Rules

Our fork version (`X.Y.Z`) follows standard semantic versioning:

- **MAJOR** version when we make incompatible API changes
- **MINOR** version when we add functionality in a backward-compatible manner
- **PATCH** version when we make backward-compatible bug fixes

### When to Bump Versions

#### For Our Fork's Changes

| Change Type | Example | Version Bump |
|------------|---------|-------------|
| Add new feature | New endpoint added | Minor: `v1.0.0` → `v1.1.0` |
| Fix bug | Bug fix in existing code | Patch: `v1.0.0` → `v1.0.1` |
| Breaking change | API signature changed | Major: `v1.0.0` → `v2.0.0` |

#### For Upstream Merges

Match or exceed the upstream version bump type:

| Upstream Change | Upstream Version | Our Fork Bump | Example |
|----------------|------------------|----------------|---------|
| Patch (bug fixes only) | `v1.0.2` → `v1.0.3` | Patch minimum | `v1.0.0` → `v1.0.1+upstream.1.0.3` |
| Minor (new features) | `v1.0.2` → `v1.1.0` | Minor minimum | `v1.0.0` → `v1.1.0+upstream.1.1.0` |
| Major (breaking changes) | `v1.0.2` → `v2.0.0` | Major minimum | `v1.0.0` → `v2.0.0+upstream.2.0.0` |

**Guidelines:**
- **Upstream patch** (bug fixes): Bump at least patch in our fork
- **Upstream minor** (new features): Bump at least minor in our fork
- **Upstream major** (breaking changes): Bump at least major in our fork
- **Our own changes**: Bump according to semantic versioning regardless of upstream
- **Combined changes**: Use the higher of the two bump types

## Automated Upstream Sync

Two GitHub Actions workflows automate the sync process:

### `sync-upstream-releases.yml`
1. Checks for new upstream releases daily (or manually)
2. Compares against the latest `upstream-sync/v*` tag to determine current version
3. Creates a PR when a new release is detected

### `tag-upstream-sync.yml`
Automatically runs when an upstream sync PR (labeled `upstream-sync`) is merged:
1. Creates an `upstream-sync/vX.Y.Z` tag to track the sync
2. This tag is used by both the sync workflow and versioning scripts

## Tagging a New Release

### Automated Tagging (Recommended)

After merging changes (including upstream sync PRs), create a new release tag:

```bash
# Via GitHub UI:
# Actions → tag-semver → Run workflow
# Select method: next

# Via GitHub CLI:
gh workflow run tag-semver.yml -f method=next -f annotation="Release notes"
```

The `next` method will:
- Analyze all commits since the last tag
- Determine the appropriate version bump (major/minor/patch) based on commit messages
- Always bump at least patch version (prevents duplicate tags)
- Append `+upstream.X.Y.Z` from the latest `upstream-sync/v*` tag
- Create and push the tag

**How it works with upstream merges:**
1. The upstream sync workflow creates a PR with a semantic commit prefix (`fix:`, `feat:`, or `feat!:`) based on the upstream version bump type
2. When merged, the `tag-upstream-sync` workflow automatically creates an `upstream-sync/vX.Y.Z` tag
3. When you run `tag-semver`, it reads this tag to add the correct `+upstream.X.Y.Z` metadata

### Manual Tagging

If you prefer to create tags manually:

```bash
# 1. Determine the appropriate version bump
# Current version: v1.1.0+upstream.1.0.2
# Upstream sync: v1.1.0 -> bump to v1.2.0

# 2. Get the current upstream version from tags
UPSTREAM_VERSION=$(git tag -l 'upstream-sync/v*' --sort=-v:refname | head -n1 | sed 's|upstream-sync/v||')

# 3. Create and push the tag
git tag -a v1.2.0+upstream.$UPSTREAM_VERSION -m "chore: release v1.2.0 based on upstream v$UPSTREAM_VERSION"
git push origin v1.2.0+upstream.$UPSTREAM_VERSION
```

## Why This Format?

- **Build metadata** (`+upstream.A.B.C`) is part of SemVer spec but ignored in version comparisons
- Our main version (`X.Y.Z`) ensures proper version ordering
- Upstream version tracking helps with:
  - Debugging issues that may be upstream-related
  - Understanding which upstream features are available
  - Planning future syncs

## Go Module Considerations

Go modules treat the build metadata (`+upstream.X.Y.Z`) as informational only:

```go
// These are treated identically by Go:
require github.com/accelleran/entrest v1.0.0+upstream.1.0.2
require github.com/accelleran/entrest v1.0.0+upstream.1.1.0

// But these are different versions:
require github.com/accelleran/entrest v1.0.0+upstream.1.0.2
require github.com/accelleran/entrest v1.1.0+upstream.1.0.2  // ✓ newer
```

This is why we always bump the base version (`X.Y.Z`) for any change.

## Manual Sync Process

If the automated workflow encounters conflicts or you want to manually sync:

1. **Check for updates:**
   ```bash
   ./.github/scripts/check-upstream-version.sh
   ```

   The script will output:
   - Current upstream version (from latest `upstream-sync/v*` tag)
   - Latest upstream version (from upstream repository)
   - Update type (`major`, `minor`, or `patch`)

2. **Map update type to semantic commit prefix:**
   - `major` → use `feat!:` (breaking change)
   - `minor` → use `feat:` (new feature)
   - `patch` → use `fix:` (bug fix)

3. **Create a sync branch from the upstream tag:**
   ```bash
   UPSTREAM_VERSION="v1.0.2"  # The version you want to sync
   git checkout -b "sync-upstream-${UPSTREAM_VERSION}" "${UPSTREAM_VERSION}"
   git push origin "sync-upstream-${UPSTREAM_VERSION}"
   ```

4. **Create a PR with the semantic commit prefix in the title:**
   ```
   {prefix}: merge upstream release vX.Y.Z
   ```

   Add the `upstream-sync` label to the PR.

5. **After merging:**
   - The `tag-upstream-sync` workflow will automatically create the `upstream-sync/vX.Y.Z` tag
   - No manual tag creation needed!

The PR will show any conflicts that need to be resolved before merging.

## FAQ

**Q: Why not use `-fork.X` instead of `+upstream.X.Y.Z`?**
A: The `-` creates a pre-release version which is considered "less than" the final release. We want our versions to be treated as stable releases.

**Q: What if we diverge significantly from upstream?**
A: Continue the versioning scheme. The upstream version in the tag shows the last point of sync, helping understand the divergence point.

**Q: Can we skip upstream versions?**
A: Yes! If upstream releases v1.1.0 and v1.2.0, and you only want v1.2.0, just merge that. The `upstream-sync/v*` tag will track it.

