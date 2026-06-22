# Git Email Export Identity

Exports are associated with an Export Identity derived from the user's normalized Git email when available, rather than from LeetCode Session data or repository remotes. Git email is stable enough for cross-machine export matching while avoiding credential-derived identifiers; if no Git email is configured, Leetgo generates a local fallback identity and tells the user.

**Considered Options**: Hash LeetCode Session data, hash Git name and email, hash repository remote, generate only a local ID.
**Consequences**: Users with different Git emails across machines may get separate Export Identities, and Leetgo must never treat the LeetCode Session as identity material.
