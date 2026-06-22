# Leetgo

Leetgo helps software engineers practice LeetCode Problems through a structured Roadmap, local workspaces, progress tracking, and gamified feedback.

## Language

### Core Domain

**Problem**:
A single LeetCode exercise identified by its LeetCode ID, title, difficulty, and category.
_Avoid_: Question, challenge, exercise

**Roadmap**:
A curated directed acyclic graph (DAG) of Problems with a specific practice goal and audience. Edges represent prerequisite relationships; solving a Problem unlocks its dependents within that Roadmap.
_Avoid_: List, curriculum, syllabus, path

**Default Roadmap**:
The Roadmap selected for new users unless they choose another one. Leetgo's Default Roadmap is `from-zero-to-hero`.
_Avoid_: Main list, beginner mode

**Roadmap Selection**:
The user-facing choice of which Roadmap to practice. During Onboarding, the recommended Roadmap is focused first and must be confirmed explicitly.
_Avoid_: Roadmap picker, mode select

**Roadmap Carousel**:
The Roadmap Selection presentation where the focused Roadmap is centered and neighboring Roadmaps are previewed to the left and right.
_Avoid_: Roadmap grid, card list

**Category**:
A grouping label for Problems that share a common algorithmic pattern or data structure (e.g., Sliding Window, Trees, Dynamic Programming).
_Avoid_: Topic, tag, subject

**Stage**:
An ordered section within a Roadmap that groups Problems into a visible progression step for the user. Stages do not unlock Problems; prerequisite edges do.
_Avoid_: Module, chapter, unit

**Unlock Path**:
The user-facing progression through a Roadmap, emphasizing Solved Problems, Available Problems, and the prerequisites blocking locked Problems.
_Avoid_: Graph drawing, dependency chart

**Dashboard**:
The main Leetgo screen after Onboarding, focused on the user's next best action while showing supporting Profile, gamification, and Roadmap context.
_Avoid_: Home page, stats page, main menu

**Next Action**:
A recommended action the user can take from the Dashboard, such as continuing an InProgress Problem, starting an Available Problem, reviewing a blocker, or submitting a solution.
_Avoid_: Task, todo, suggestion

**Roadmap Detail**:
The drill-down view of a Roadmap, focused on its Unlock Path and Stage progress.
_Avoid_: Roadmap page, graph screen

**Stage Detail**:
The drill-down view of one Stage, focused on its Problems, completion, and blockers.
_Avoid_: Category page, module view

**Problem Detail**:
The drill-down view of one Problem, focused on Start, local testing, Submission, and Solve Log history.
_Avoid_: Problem page, challenge view

**Difficulty**:
A Problem's complexity tier: Easy, Medium, or Hard. Inherited from LeetCode's classification.
_Avoid_: Level, rank

### Workspace

**Onboarding**:
The first-run flow where a user creates their Profile, chooses practice preferences, chooses whether to enable Git Export backup, and confirms their initial Roadmap.
_Avoid_: Setup wizard, install flow

**Profile**:
A local representation of the person using Leetgo, including the name Leetgo uses to address them.
_Avoid_: Account, login, Git user

**Display Name**:
The user-chosen name shown in Leetgo's interface for a Profile.
_Avoid_: Username, handle, Git name

**Practice Preferences**:
The Profile-level choices that shape a user's Leetgo experience, including selected Roadmap, Language, Workspace, Theme, and Git Export backup preference.
_Avoid_: Settings, options, config knobs

**Theme**:
The visual style used by Leetgo's TUI. Leetgo supports RPG skill tree, clean productivity, and cyber dashboard Themes, with RPG skill tree as the default; ambient motion is Theme-specific.
_Avoid_: Skin, color scheme

**Workspace**:
The user-owned area where generated Problem files live.
_Avoid_: Project dir, output dir

**Stub**:
The generated source file for a Problem in the user's chosen language.
_Avoid_: Template, skeleton, boilerplate

**TestSuite**:
The generated test file for a Problem.
_Avoid_: Tests, test file, spec

**Language**:
A programming language supported for Stub generation, TestSuite generation, and LeetCode Submissions.
_Avoid_: Runtime, template type

### Progress & Gamification

**Attempt**:
A recorded instance of a user working on a Problem: timestamp, duration, pass/fail result from local tests, and optional self-reported difficulty rating.
_Avoid_: Try, submission, run

**Status**:
A Problem's practice state for the user: Locked, Available, InProgress, or Solved. Locked and Available depend on the selected Roadmap; Solved carries across Roadmaps.
_Avoid_: State, phase

**Start**:
The act of beginning work on an Available Problem.
_Avoid_: Open, generate, launch

**Solve**:
The act of completing a Problem well enough to mark it Solved for the user across all Roadmaps.
_Avoid_: Finish, complete, close

**XP**:
Experience points earned by solving Problems.
_Avoid_: Points, score

**Level**:
A tier derived from cumulative XP.
_Avoid_: Rank, tier, grade

**Streak**:
Consecutive calendar days with at least one solved Problem. Resets to zero on a missed day.
_Avoid_: Chain, run

**Achievement**:
A milestone unlocked by user progress.
_Avoid_: Badge, medal, trophy

**Weakness**:
A Category where the user's Attempts show repeated failure or high self-reported difficulty.
_Avoid_: Gap, deficiency, blind spot

### Integration

**Export**:
A portable snapshot of the user's Leetgo data intended for backup, sharing, or transfer to another environment.
_Avoid_: Dump, backup file

**Git Export**:
An Export written to a user-owned Git repository.
_Avoid_: GitHub sync, repo backup

**Export Identity**:
A privacy-preserving identifier used to associate Exports from the same user profile. It is derived from the user's normalized Git email when available.
_Avoid_: Session hash, account ID, login

**Submission**:
Sending a user's solution to LeetCode's judge for official validation.
_Avoid_: Upload, push, send

**Solve Log**:
A recorded result and optional note from a Submission, intended to preserve the user's solving history.
_Avoid_: Post, journal entry, publication

**Session**:
A user's authenticated LeetCode access used for Submissions.
_Avoid_: Login, auth, connection
