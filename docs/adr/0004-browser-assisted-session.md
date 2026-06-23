# Chromium-Assisted LeetCode Session

LeetCode submission uses a Chromium-assisted Session setup: `leetgo auth` launches Chrome or another Chromium-based browser with a Leetgo-owned browser profile, the user logs in through LeetCode directly, and Leetgo extracts the CSRF token and LeetCode Session cookie through Chrome DevTools Protocol.

LeetCode does not provide a stable public OAuth integration for this CLI use case. Chromium-assisted Session extraction keeps Leetgo from storing usernames or passwords while avoiding the brittle manual cookie-copy flow.

**Considered Options**: Environment variable Session, manual config file entry, browser-assisted Session callback, external LeetCode CLI provider, Chromium DevTools Protocol cookie extraction.
**Consequences**: `leetgo auth` requires Chrome or another Chromium-based browser on Linux/Windows. Firefox and Safari are unsupported initially. Users may still need to reconnect their Session periodically, and LeetCode endpoint changes can still break Submissions.
