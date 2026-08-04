---
name: pr-review
description: Standard process for reviewing, merging, and releasing PRs in the terraform-provider-portkey repo. Use when reviewing a PR, merging to main, updating the CHANGELOG, or cutting a release. Covers the local acceptance-test run (TF_ACC) that CI skips.
---

## Phase 1: Initial Assessment

```bash
# View open PRs
gh pr list

# Checkout the PR locally
gh pr checkout <PR-number>
```

### CHECKPOINT 1 - Understand the PR:
- Read the PR description and linked issues
- Ask yourself: *What problem does this solve? Is this functionality we need?*
- For new features: verify we don't already cover this functionality elsewhere

---

## Phase 2: Code Review

Run these checks locally:

```bash
# Build the provider
go build ./...

# Run static analysis
go vet ./...
gofmt -d .

# Run linter (note: may have known issues with golangci-lint v2)
golangci-lint run
```

### 2.0 Run Acceptance Tests (REQUIRED)

**IMPORTANT:** CI does NOT run acceptance tests (they're skipped without `TF_ACC`). You MUST run them locally.

```bash
# Load environment variables and run acceptance tests
set -a && source .env && set +a && \
  TF_ACC=1 TF_ACC_TERRAFORM_PATH=$(which terraform) \
  go test ./internal/provider -v -timeout 30m

# Or run tests for specific resources
set -a && source .env && set +a && \
  TF_ACC=1 TF_ACC_TERRAFORM_PATH=$(which terraform) \
  go test ./internal/provider -v -run TestAccPromptPartial -timeout 10m
```

**Do not use a bare `source .env`.** It sets shell variables without exporting
them, so `PORTKEY_API_KEY` never reaches the test process, `testAccPreCheck`
skips every acceptance test, and the run reports `ok` having tested nothing —
the exact failure this section exists to prevent. `set -a` exports them.
`TF_ACC_TERRAFORM_PATH` is needed when the `terraform` binary isn't on the
path the test harness searches.

Sanity-check that tests actually ran: the output must contain `--- PASS:` lines
for the tests you targeted. A run that prints only `ok ... 0.5s`, or `--- SKIP:`,
means the credentials never arrived.

The `.env` file contains:
- `PORTKEY_API_KEY` - Org-level API key for testing
- `TEST_WORKSPACE_ID` - Workspace UUID for tests requiring workspace_id
- `TEST_COLLECTION_ID`, `TEST_VIRTUAL_KEY`, etc. - Other test fixtures

Tests use helper functions like `getTestWorkspaceID()` to read these values.

### CHECKPOINT 2 - Evaluate code quality against these criteria:

### 2.1 Test Quality (No Malicious Compliance)
- [ ] Tests cover the full CRUD lifecycle (Create, Read, Update, Delete)
- [ ] Tests verify import functionality works
- [ ] Tests check actual attribute values, not just existence
- [ ] Tests cover edge cases (updates, deletions, error conditions)
- [ ] Tests are meaningful, not just "passes but tests nothing"

Example of a **good** test check:
```go
resource.TestCheckResourceAttr("portkey_resource.test", "name", "expected-value"),
```

Example of **malicious compliance** (avoid):
```go
// Only checks attribute exists, not its value
resource.TestCheckResourceAttrSet("portkey_resource.test", "id"),
```

### 2.2 Impact on Existing Code
- [ ] Changes are additive (new files) vs. modifying existing logic
- [ ] If modifying existing code: understand why and verify no regressions
- [ ] Check `provider.go` changes - should only add new registrations
- [ ] Verify no unrelated changes snuck in

### 2.3 Code Style Consistency
- [ ] File naming follows pattern: `{resource}_resource.go`, `{resource}_data_source.go`
- [ ] Schema definitions match existing resources (descriptions, validators, plan modifiers)
- [ ] Error handling is consistent with other resources
- [ ] API client methods follow existing patterns in `client/client.go`

### 2.4 Documentation
- [ ] Resource docs exist in `docs/resources/` or `docs/data-sources/`
- [ ] Examples are provided
- [ ] RESOURCE_MATRIX.md is updated (if applicable)

---

## Phase 3: Provide Feedback

**If changes are needed:**

```bash
gh pr review <PR-number> --request-changes --body "Your detailed feedback"
```

Be specific about:
- What needs to change and why
- Provide code examples when helpful
- Reference existing code patterns to follow

**If PR is good:**

```bash
gh pr review <PR-number> --approve --body "Your approval message"
```

### CHECKPOINT 3 - Before approving, verify:
- [ ] CI checks are passing (or failures are pre-existing/unrelated)
- [ ] **You've run acceptance tests locally** (CI only runs unit tests, not acceptance tests!)
- [ ] Code compiles without errors
- [ ] You understand what the code does

**WARNING:** A green CI does NOT mean tests pass. CI skips acceptance tests (`TF_ACC` not set). Always run them yourself using the exact invocation in section 2.0, and confirm you see `--- PASS:` lines rather than a bare `ok`.

---

## Phase 4: Merge

```bash
# Merge the PR
gh pr merge <PR-number> --merge

# Pull changes to local main
git checkout main && git pull
```

---

## Phase 5: Update CHANGELOG

Edit `CHANGELOG.md`:
1. Add entry under `[Unreleased]` describing what was added/changed/fixed
2. Follow the existing format and categorization (Added, Changed, Fixed, etc.)

**`main` is protected — this cannot be committed directly.** See Phase 6 for
the branch-and-PR flow; the CHANGELOG edit has to ride in a PR like anything
else. If you're releasing straight away, skip this phase and fold the entry
into the release PR instead of raising two.

Before writing the entry, check what actually landed since the last tag —
merged PRs routinely miss their CHANGELOG update, and the release is the last
chance to catch them:

```bash
git log --oneline <last-tag>..HEAD --merges
```

---

## Phase 6: Release

When ready to release (can batch multiple PRs).

**`main` is protected by the org-level ruleset "Primary Rules"
(`deletion`, `non_fast_forward`, `pull_request`), so `git push origin main`
is rejected.** Every change to `main`, including a one-line CHANGELOG edit,
goes through a PR. The ruleset requires:

- 1 approving review, and **`bypass_actors` is empty** — repo admins get no override
- `require_last_push_approval` — any push after approval dismisses it
- `dismiss_stale_reviews_on_push` and `required_review_thread_resolution`

**You cannot approve your own PR.** GitHub blocks self-approval, so a release
PR you author needs a *second maintainer* to approve it. Plan for that
handoff — don't start a release you can't finish, and tell the user up front
that the merge will block on someone else.

```bash
# 1. Branch first — naming convention is chore/release-vX.Y.Z (see PR #53)
git checkout main && git pull
git checkout -b chore/release-vX.Y.Z

# 2. Update CHANGELOG.md
#    - Change [Unreleased] to [X.Y.Z] - YYYY-MM-DD
#    - Add new empty [Unreleased] section
#    - Update comparison links at the bottom (add the new version, and
#      repoint [Unreleased] to compare/vX.Y.Z...HEAD)

# 3. Commit, push, open the PR
git add CHANGELOG.md
git commit -m "chore: update CHANGELOG for vX.Y.Z"
git push -u origin chore/release-vX.Y.Z
gh pr create --title "chore: update CHANGELOG for vX.Y.Z" --body "..."

# 4. Wait for CI, then STOP and get an approval from another maintainer.
#    gh pr view <PR> --json mergeStateStatus,reviewDecision
#    BLOCKED / REVIEW_REQUIRED means the approval is still outstanding.
gh pr merge <PR-number> --merge

# 5. Tag only AFTER the CHANGELOG PR is merged, from updated main.
#    The ruleset targets branches, not tags, so pushing a tag is a
#    direct push and works fine.
git checkout main && git pull
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z

# 6. Monitor release
gh run watch $(gh run list --workflow=release.yml --limit 1 --json databaseId -q '.[0].databaseId') --exit-status

# 7. Verify
gh release view vX.Y.Z
```

Tagging a commit whose CHANGELOG PR hasn't merged ships a release whose notes
don't match its contents, and the tag can't be moved afterward
(`non_fast_forward`). Confirm `git log --oneline -1` on `main` is the merge
commit before step 5.

---

## Troubleshooting

### CI Lint Failures
The `golangci-lint` configuration may have compatibility issues with certain Go versions. If lint fails in CI but the PR code is correct:
- Verify the lint errors are pre-existing (not introduced by the PR)
- Check if errors are in files not touched by the PR
- The lint config (`.golangci.yml`) may need updates when Go version changes

### Force Push After Rebase
If rebasing a PR branch and `git push --force-with-lease` fails with "stale info":
```bash
git push --force  # Use with caution, only on PR branches
```

### PR Stuck at `BLOCKED` / `REVIEW_REQUIRED`
The org ruleset requires one approving review and has no bypass actors, so
this is normal for a PR you authored yourself — GitHub won't let you approve
your own. Get another maintainer to approve; there is no admin override.
Check what's actually outstanding with:
```bash
gh pr view <PR-number> --json mergeStateStatus,reviewDecision
gh api repos/Portkey-AI/terraform-provider-portkey/rulesets/15003178 \
  --jq '.rules[] | select(.type=="pull_request") | .parameters'
```
`BLOCKED` with `reviewDecision=APPROVED` means something else is pending —
usually an unresolved review thread (`required_review_thread_resolution`), or
an approval dismissed by a later push (`require_last_push_approval`).

### `no checks reported on the '<branch>' branch`
CI takes up to a minute to register on a freshly pushed branch. Re-check
before concluding the workflow didn't trigger:
```bash
gh run list --branch <branch> --limit 5
```

### CI Flakiness
If CI fails intermittently:
- Re-run the workflow: `gh run rerun <run-id>`
- Check if it's a timing/race condition in tests
- Check if it's an API rate limit issue

### Contributor Can't Push Fixes
If a contributor needs help making changes and "Allow edits from maintainers" is enabled:
```bash
# Add their fork as remote
git remote add contributor-fork https://github.com/CONTRIBUTOR/terraform-provider-portkey.git

# Push fixes to their branch
git push contributor-fork HEAD:their-branch-name
```

---

## Review Mindset

Every PR review is a learning opportunity. After each review session, take a moment to reflect:

- Did you learn a new pattern or technique from the contributor's code?
- Did you discover an edge case you hadn't considered before?
- Did the discussion reveal gaps in documentation or testing patterns?
- Could this PR's approach improve other parts of the codebase?

Good feedback flows both ways. Stay curious, ask questions when something is unclear, and remember that even experienced contributors can learn something new from every code review.
