## Sync Upstream Release ${LATEST_UPSTREAM}

This PR merges changes from the upstream [lrstanley/entrest](https://github.com/lrstanley/entrest) repository.

### Version Information
- **Previous upstream version:** ${CURRENT_UPSTREAM}
- **New upstream version:** ${LATEST_UPSTREAM}
- **Update type:** ${UPDATE_TYPE}
- **Upstream release notes:** https://github.com/lrstanley/entrest/releases/tag/${LATEST_UPSTREAM}

### What Happens After Merge?
When this PR is merged, the `tag-upstream-sync` workflow will automatically create an `upstream-sync/${LATEST_UPSTREAM}` tracking tag to record that we've synced to this upstream version.

No manual action is required! 🎉

See [VERSIONING.md](./VERSIONING.md) for details on our versioning strategy and how to create release tags.

### Review Checklist
- [ ] Review upstream changes for compatibility
- [ ] Test with existing integrations
- [ ] Update documentation if needed
- [ ] Tag new release after merge

---
*Automatically created by sync-upstream-releases workflow*

