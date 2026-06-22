# Bundled Curated Problem Catalog

The problem graph ships as an embedded YAML file inside the binary, not fetched from LeetCode's API at runtime.

LeetCode's API is unofficial, unstable, and rate-limited. A bundled catalog gives us a known-good, versioned set of problems with hand-curated prerequisites and categories. Users get a consistent experience offline and across versions.

**Considered Options**: Scrape LeetCode API, let users define their own graph, ship a curated YAML.
**Consequences**: New LeetCode problems won't appear until a new release. Users cannot add custom problems in v1.
