# Roadmap File Per Roadmap

Bundled Roadmaps are stored as one YAML file per Roadmap, with each file owning its metadata, ordered Stages, Problems, and prerequisite edges. We chose this over a single multi-roadmap catalog because Roadmaps are curated products that should be reviewed, versioned, and evolved independently.

**Considered Options**: Single catalog file, shared Problem catalog plus Roadmap reference files, one file per Roadmap.
**Consequences**: The same LeetCode Problem may be duplicated across Roadmap files, and catalog loading must select a Default Roadmap when no Roadmap is configured.
