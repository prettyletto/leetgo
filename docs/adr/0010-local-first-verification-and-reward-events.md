# Solved-Gated Progression and Reward Events

Leetgo treats local TestSuite pass as `Verified`, a local confidence state that can earn local XP but does not unlock dependent Problems. Roadmap progression is gated by `Solved`: either an Accepted LeetCode Submission (`Accepted Solve`) or an explicit user confirmation (`Manual Solve`). Manual Solve unlocks dependent Problems but earns no XP unless the user later upgrades it with an Accepted Submission.

**Considered Options**: Make local `Verified` unlock dependents, make Accepted Submission the only valid Solve, allow Manual Solve as an unlock-only override, award XP for Manual Solve.
**Consequences**: Dashboard, Problem Detail, CLI output, and Roadmap unlocking must distinguish `Verified`, `Manual Solve`, and `Accepted Solve`. Verified Problems become strong Next Actions for Submission or Manual Solve. Manual Solve needs confirmation and Practice Log provenance. Existing Reward Events remain idempotent, and Accepted Submission can later upgrade a Manual Solve for eligible XP.
