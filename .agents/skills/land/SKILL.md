---
name: land
description: >-
  Land requested changes in the Burrow repository (Go module
  github.com/raiden0456/gonteractive) on its GitHub main branch and verify
  Docker publication. Invoke only when the user explicitly requests landing
  or merging the relevant change, including an explicit /land invocation,
  not for review, preparation, passing checks, or skill installation.
disable-model-invocation: true
metadata:
  delta-action: land
---

# Land Burrow changes

## Scope and authorization

Apply only to this repository: the Go module is
`github.com/raiden0456/gonteractive` in `go.mod`, and the canonical GitHub
repository is `Raiden0456/Burrow`. The historical
`Raiden0456/gonteractive` remote URL is acceptable only when GitHub confirms
it resolves to that canonical repository. Do not apply this workflow to
unrelated repositories.

An explicit invocation supplies authorization to land the relevant change
using this workflow, including the Docker publication triggered by the push.
Do not ask for the same merge permission again or respond to `/land` by
waiting for another landing request. Installation approval alone is not
landing authorization. Stop for unresolved scope, conflicts, failed checks,
changed destination requirements, or unsafe operations.

The intended destination is `origin/main`, using a normal fast-forward push.
Do not force-push, rewrite existing commits, bypass protections, modify
repository permissions, or update a running deployment.

## Preflight

1. Read applicable repository instructions, especially `CLAUDE.md`, and any
   contribution policies or submission templates added since this skill was
   written. Honor applicable signing, contributor agreement, review,
   documentation, and submission-text requirements even if GitHub does not
   enforce them. If a requirement needs human-authored text, obtain that text
   rather than substituting generated text with human approval.
2. Inspect working-tree status, staged and unstaged diffs, untracked files,
   current branch, and commits ahead of the destination. Identify the exact
   changes requested for landing. Preserve existing commits and unrelated
   work. Do not stage the entire working tree indiscriminately. If the
   requested and unrelated changes cannot be separated safely, stop and
   briefly ask which changes belong.
3. Verify installed prerequisites against `go.mod`, the `packageManager`
   field in `web/package.json`, and the build stages in `Dockerfile`.
   The researched baseline uses Go 1.27.1 or compatible newer Go, Node.js 24,
   pnpm 11.13.0, Make, Git, and the real GitHub CLI `gh`. Check their
   availability and `gh auth status` without printing credential values.
   Do not silently substitute an unverified wrapper or incompatible runtime.
4. Inspect Git remotes before publishing. Use the source remote `origin`,
   never the `local` backlink. Resolve its repository identity through
   GitHub and require `Raiden0456/Burrow`. Do not silently change remote
   URLs.
5. Read the live default branch, write permissions, branch protection, and
   effective rules for `main`, for example through `gh api` repository,
   branch, protection, and branch-rules endpoints. A confirmed unprotected
   branch is different from an authentication failure or an unverifiable
   setting. The researched configuration has `main` as the default,
   permits direct pushes, and has no required pre-landing GitHub checks;
   this is a baseline to recheck, not a permanent exemption.
6. If policy now requires a PR, review, merge queue, signing, or status
   checks unavailable before this direct push, stop and explain the unmet
   requirement. Do not bypass it or claim success from a prepared branch.
   Obtain a decision on the required alternate process rather than
   inventing one.
7. Fetch the latest destination with `git fetch origin main` and record its
   exact SHA. Compare it with the proposed commit history. If there is
   nothing left to land, verify that the intended change is present at the
   destination and inspect its relevant Docker run before reporting the
   existing result.

## Prepare and verify the exact change

Use the following project commands from the repository root, in this order.
Do not run the build concurrently with tests: the Go binary embeds
`web/dist`, which a fresh checkout may not contain.

- For changes to migrations or query inputs, run `make generate` and inspect
  the generated diff. Never edit `internal/store/db` manually. Preserve
  existing migrations; schema updates require a new migration.
  Sources: `Makefile`, target `generate` (lines 20–21), and the store
  conventions in `CLAUDE.md`.
- Run `make build`. Its `web` prerequisite installs frozen frontend
  dependencies and builds the embedded SPA before compiling the Go binary.
  Sources: `Makefile`, targets `web` and `build` (lines 10–14);
  `web/package.json`, script `build`; `web/embed.go`, the `all:dist` embed.
- Run `make test`. This executes all Go tests with the race detector,
  frontend lint, and the TypeScript project check.
  Source: `Makefile`, target `test` (lines 16–18).
- Run `git diff --check` and inspect tracked changes left by generation or
  verification. If commands changed source or generated tracked files,
  include only the relevant verified changes and repeat affected checks.
  Do not treat earlier checks on another source version as sufficient.

For a strictly documentation-only change, with no executable behavior,
schema, dependencies, CI, or build configuration changed, local verification
may be limited to reviewing the exact diff, references, skill frontmatter
where applicable, and `git diff --check`. This exemption does not cover
landing a documentation file together with an application feature.

All applicable local checks and all live required pre-landing GitHub checks
must pass for the exact changes being landed. Pending, failing, missing,
cancelled, skipped when required, or unverifiable required checks are not
success. If the embedded frontend is missing, build it with `make build`
and rerun `make test`; do not replace the full test target with a subset.
If a prerequisite is missing, report the blocker rather than installing
system tools or silently weakening verification.

Inspect `git status` after building. Exclude generated `bin/`, `web/dist/`
including a newly created `.gitkeep`, local databases, `.env`, credentials,
and unrelated files from new commits. These paths are build/local artifacts;
see `Makefile` and `.gitignore`. Do not delete unrelated or local data.

Commit any relevant uncommitted changes with an explicit file selection
and a concise descriptive message, following current repository conventions.
Keep existing commits intact. Use `GIT_EDITOR=true git commit -m "…"` so
the command cannot open an editor. If the newly installed landing skill is
part of the requested change, include it explicitly.

## Integrate and push

1. Ensure relevant source changes are committed before merging destination
   history. Do not discard, stash, or overwrite unrelated work to proceed.
   Use a task branch if needed to avoid disturbing another local branch;
   never reset or rewrite a shared branch.
2. If the fetched `origin/main` is already an ancestor of the proposed
   HEAD, retain the existing history. Otherwise, when the histories are
   related and the landing scope remains clear, incorporate it with
   `GIT_EDITOR=true git merge --no-edit origin/main`. Do not use an
   interactive rebase or force-push.
3. **Stop on every conflict. Do not resolve any conflict automatically,
   even if the intended fix appears obvious.** Briefly state the failed
   operation and conflicting paths, say that landing has not completed,
   and wait for the user's decision. Preserve the state; do not silently
   abort, reset, or choose either side.
4. After incorporating any new destination changes, rerun verification
   on the resulting combined tree. Verify all applicable requirements
   against that version, not only the original feature commit.
5. Immediately before publishing, recheck the destination SHA, rules, and
   required checks. If `main` advanced, fetch and integrate the new tip,
   stop on any conflict, and repeat the affected verification. If
   requirements or scope changed, stop rather than bypassing them.
6. Record the final commit SHA, then use the normal push
   `git push origin HEAD:refs/heads/main`. A non-fast-forward rejection
   is not success: refresh the destination and repeat integration and
   verification, stopping if the destination keeps moving or intent
   becomes ambiguous. Never use `--force`, `--force-with-lease`, an admin
   bypass, or a destructive reset.

The publication behavior is defined in `.github/workflows/docker.yml`:
pushes to `main` trigger the `docker` workflow; its two build jobs publish
linux/amd64 and linux/arm64 images, and its dependent `merge` job publishes
the manifest tags. This is a post-push publication workflow, not a
pre-merge test workflow. Its image name is derived from the canonical
`GITHUB_REPOSITORY`, currently `ghcr.io/raiden0456/burrow`; prefer this
implementation over stale image names in documentation. Do not trigger a
separate manual publication or deploy containers as part of landing.

## Confirm destination and publication

1. Read the remote `main` SHA after the push. Verify the exact pushed
   commit is the destination tip, or is included in a newer tip if another
   contributor advanced it. Do not infer this from local branch status
   alone.
2. Find the `docker` push run for the exact landed SHA using `gh run list`
   filtered by commit and workflow. Allow time for GitHub to create the
   run; do not use the green result of a previous commit. Wait with
   `gh run watch <run-id> --repo Raiden0456/Burrow --exit-status`, subject
   to the terminal's time limit. Inspect the final job results and require
   both architecture builds and the dependent manifest job to succeed.
   Source for the workflow, trigger, dependency, matrix, and final manifest
   inspection: `.github/workflows/docker.yml`, especially lines 3–7,
   13–23, 47–56, and 72–112.
3. If a terminal wait times out, inspect the same run rather than starting
   another publication. A missing, pending, failing, or unverifiable run
   must not be described as a successful image publication.
4. State the destination branch, landed commit, checks actually run,
   Docker run link and conclusion, and any remaining local changes.
   If blocked before pushing, explicitly say the change has not landed.
   If the code reached `main` but Docker publication failed or is still
   pending, clearly distinguish **code landed** from **image not verified
   or not published**. Do not claim the running application has updated.
5. Do not automatically revert a landed commit, delete branches, clean
   untracked files, publish tags, or restart a deployment. Those are
   separate operations requiring their own scope and authorization.
