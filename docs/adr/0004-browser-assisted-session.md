# Browser-Assisted LeetCode Session

LeetCode submission uses a browser-assisted Session setup: the user logs in through the browser, then gives leetgo the CSRF token and LeetCode session value used for Submissions.

LeetCode does not provide a stable public OAuth integration for this CLI use case. A browser-assisted Session keeps the first version honest about the unofficial integration while avoiding storing credentials or passwords.

**Considered Options**: Environment variable Session, manual config file entry, browser-assisted Session callback.
**Consequences**: Users may need to refresh their Session periodically, and LeetCode endpoint changes can still break Submissions.
