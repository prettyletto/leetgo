# Local-First Verification and Reward Events

Leetgo treats local TestSuite pass as the primary success state (`Verified`) and LeetCode Accepted Submission as optional external confirmation (`Solved`). This makes the app useful offline and prevents an unstable browser-session LeetCode integration from blocking Roadmap progression or XP awards. XP is awarded through idempotent Reward Events: 70% on first local Verify, 30% on first Accepted Submission, 0% on manual Solve.

**Considered Options**: Make LeetCode Accepted the only valid solve condition, split rewards 50/50, make local Verify worth 100% with no Submission XP, award XP on every test pass.
**Consequences**: Existing `Solved` Problems must be migrated with legacy Reward Event guards to prevent duplicate XP. Dashboard, Problem Detail, CLI output, and Roadmap unlocking all need to distinguish Verified from Solved. LeetCode Submission becomes best-effort, not a core dependency.
