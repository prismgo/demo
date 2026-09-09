---
name: code-review
description: Review committed or working-tree changes since a fixed point along three default axes — Correctness, Concurrency, Performance — each a parallel sub-agent. Standards and Security are opt-in. Use for branch, PR, or WIP review.
---

Review changes against a fixed point the user supplies, along three axes:

- **Correctness** — does the diff's logic actually do what it appears to intend?
- **Concurrency** — does the diff introduce a race, deadlock, or other concurrency hazard?
- **Performance** — does the diff introduce an avoidable performance cost?

Two more axes exist but only run if the user explicitly asks for them (by name, or "full review"):

- **Standards** — does the code conform to this repo's documented coding standards?
- **Security** — does the diff introduce or fail to close a security risk?

Each active axis runs as its own **parallel sub-agent** so they don't pollute each other's context, then this skill aggregates their findings.

## Process

### 1. Pin the fixed point and target

Whatever the user said is the fixed point — a commit SHA, branch name, tag, `main`, `HEAD~5`, etc. If they didn't specify one, ask for it. Resolve it once with `git rev-parse <fixed-point>`.

Choose the target from the request:

- **Committed target** (default for branches and PRs): use `git diff <fixed-point>...HEAD` and note commits with `git log <fixed-point>..HEAD --oneline`.
- **Working-tree/WIP target** (when explicitly requested): when the project provides a WIP patch helper, prefer it; PrismGo uses `bash demo/.agents/skills/dev-harness-lite/scripts/wipDiff.sh <repo> <fixed-point> /tmp/<name>.patch`. Otherwise use `git diff <fixed-point>` for tracked changes, list new files with `git ls-files --others --exclude-standard`, and include each new file with `git diff --no-index -- /dev/null <file>` (exit 1 means a non-empty patch, not failure). This mode includes committed, staged, unstaged, and untracked work without changing the real index.

Confirm the complete selected change set is non-empty before spawning reviewers. A bad ref or empty combined diff fails here. Save a large combined patch under `/tmp` and pass its path instead of copying it through the parent context.

For an incremental WIP re-review, save each combined patch. Give reviewers `diff -u <previous.patch> <current.patch>` plus the current affected files; they review only changes since the last reviewed patch while retaining the original fixed point. Do not silently fall back to `...HEAD`, because that drops uncommitted work.

### 2. Identify which axes run

Default set: **Correctness, Concurrency, Performance**. Skip straight to step 3 for these.

Only if the user explicitly asked for **Standards** and/or **Security**, add them and do the matching setup below before spawning sub-agents.

#### 2a. Identify the standards sources (only if Standards was requested)

Anything in the repo that documents how code should be written, such as `CODING_STANDARDS.md` or `CONTRIBUTING.md`.

On top of whatever the repo documents, the Standards axis always carries the **smell baseline** below — a fixed set of Fowler code smells (_Refactoring_, ch.3) that applies even when a repo documents nothing. Two rules bind it:

- **The repo overrides.** A documented repo standard always wins; where it endorses something the baseline would flag, suppress the smell.
- **Always a judgement call.** Each smell is a labelled heuristic ("possible Feature Envy"), never a hard violation — and, like any standard here, skip anything tooling already enforces.

Each smell reads *what it is* → *how to fix*; match it against the diff:

- **Mysterious Name** — a function, variable, or type whose name doesn't reveal what it does or holds. → rename it; if no honest name comes, the design's murky.
- **Duplicated Code** — the same logic shape appears in more than one hunk or file in the change. → extract the shared shape, call it from both.
- **Feature Envy** — a method that reaches into another object's data more than its own. → move the method onto the data it envies.
- **Data Clumps** — the same few fields or params keep travelling together (a type wanting to be born). → bundle them into one type, pass that.
- **Primitive Obsession** — a primitive or string standing in for a domain concept that deserves its own type. → give the concept its own small type.
- **Repeated Switches** — the same `switch`/`if`-cascade on the same type recurs across the change. → replace with polymorphism, or one map both sites share.
- **Shotgun Surgery** — one logical change forces scattered edits across many files in the diff. → gather what changes together into one module.
- **Divergent Change** — one file or module is edited for several unrelated reasons. → split so each module changes for one reason.
- **Speculative Generality** — abstraction, parameters, or hooks added for needs the spec doesn't have. → delete it; inline back until a real need shows.
- **Message Chains** — long `a.b().c().d()` navigation the caller shouldn't depend on. → hide the walk behind one method on the first object.
- **Middle Man** — a class or function that mostly just delegates onward. → cut it, call the real target direct.
- **Refused Bequest** — a subclass or implementer that ignores or overrides most of what it inherits. → drop the inheritance, use composition.

### 3. Spawn sub-agents in parallel

Always spawn the three default sub-agents. Spawn the other two only for whichever of Standards/Security the user asked for in step 2.

**Correctness sub-agent prompt** — include:

- The selected target mode, full diff command(s) or saved patch path, and commit list.
- The brief: "Report every place the diff's logic doesn't do what it's clearly trying to do: wrong conditionals or boundary conditions, off-by-one, unhandled edge cases (nil/empty/zero/error paths), incorrect state left behind after a partial failure, silently dropped or swallowed errors, wrong operator/comparison, wrong order of operations. For each: quote the hunk, state the concrete input or sequence that triggers it, and the expected vs. actual behavior. Distinguish confirmed bugs from suspected-but-unverified. Skip style-only nits. Under 400 words."

**Concurrency sub-agent prompt** — include:

- The selected target mode, full diff command(s) or saved patch path, and commit list.
- The brief: "Report every place the diff introduces a concurrency hazard: shared mutable state read or written without synchronization, check-then-act races (TOCTOU), potential deadlock from lock ordering across goroutines/threads, leaked goroutines/threads/timers/subscriptions, missing cancellation or timeout propagation, unsafe reuse of a connection/resource across concurrent callers, double-close or use-after-close. For each: quote the hunk and describe the interleaving that breaks it — which two paths race, and in what order. If the diff touches no concurrent code, say so and skip. Under 400 words."

**Performance sub-agent prompt** — include:

- The selected target mode, full diff command(s) or saved patch path, and commit list.
- The brief: "Report every place the diff introduces an avoidable performance cost: N+1 queries, unbounded or unindexed queries (missing LIMIT/pagination), avoidable superlinear complexity where linear is achievable, unbounded memory growth (unbounded caches/slices/goroutines), blocking or synchronous I/O inside a loop or hot path, missing batching for repeated work that could be batched. For each: quote the hunk, state the growth factor or hot-path trigger, and the scale at which it becomes a real problem. Distinguish measured/obvious costs from speculative micro-optimizations — skip the latter. Under 400 words."

**Standards sub-agent prompt** (opt-in) — include:

- The selected target mode, full diff command(s) or saved patch path, and commit list.
- The list of standards-source files you found in step 2a, **plus the smell baseline from step 2a** pasted in full — the sub-agent has no other access to it.
- The brief: "Report — per file/hunk where relevant — (a) every place the diff violates a documented standard: cite the standard (file + the rule); and (b) any baseline smell you spot: name it and quote the hunk. Distinguish hard violations from judgement calls — documented-standard breaches can be hard, but baseline smells are always judgement calls, and a documented repo standard overrides the baseline. Skip anything tooling enforces. Under 400 words."

**Security sub-agent prompt** (opt-in) — include:

- The selected target mode, full diff command(s) or saved patch path, and commit list.
- The brief: "Report every place the diff introduces or fails to close a security risk: injection (SQL/command/template/XSS), missing authn/authz checks or IDOR, sensitive data (secrets/tokens/PII) logged, exposed, or hardcoded, insecure deserialization, path traversal/SSRF, missing input validation or sanitization at a trust boundary, weak or missing crypto/randomness, unsafe eval/exec. For each: quote the hunk, name the vulnerability class, and give a concrete exploit scenario — what input or actor triggers it. Under 400 words."

### 4. Aggregate

Present each report that ran under its own heading — `## Correctness`, `## Concurrency`, `## Performance`, plus `## Standards` / `## Security` if requested — verbatim or lightly cleaned, in default-then-opt-in order. Do **not** merge or rerank findings across axes — they are deliberately separate (see _Why these axes_). Omit a heading entirely if that sub-agent found nothing, rather than printing "no findings."

End with a one-line summary: total findings per axis, and the worst issue _within each axis_ (if any). Don't pick a single winner across axes — that's the reranking the separation exists to prevent.

## Why these axes

A change can pass one axis and fail another — that's exactly the case a single-pass review misses:

- Code that's correct single-threaded but races under concurrent load → **Correctness pass, Concurrency fail.**
- Code that's correct and safe but does an unbounded query per row → **Correctness pass, Performance fail.**
- Code that works correctly for the happy path but trusts unsanitized input → **Correctness pass, Security fail** (opt-in).

Correctness, Concurrency, and Performance are default because they catch the functional bugs that break a system regardless of whether a standards doc exists. Standards needs extra setup (a docs lookup) and Security needs a dedicated threat-modeling pass — both stay opt-in so a routine review doesn't pay that cost, but are one word away (`"also check security"`) when the change warrants it.

Reporting each axis separately, from independent sub-agents that don't see each other's framing, stops any one axis from masking the others.
