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

**Roadmap Time Estimate**:
A broad commitment range for completing a Roadmap at an assumed practice pace.
_Avoid_: Deadline, completion date, exact ETA

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

**Practice Focus**:
The specific pattern or skill a Problem is intended to train, usually narrower than its Category.
_Avoid_: Topic, tag, lesson

**Stage**:
An ordered section within a Roadmap that groups Problems into a visible progression step for the user. Stages do not unlock Problems; prerequisite edges do.
_Avoid_: Module, chapter, unit

**Unlock Path**:
The user-facing progression through a Roadmap, emphasizing Solved Problems, Available Problems, and the prerequisites blocking locked Problems.
_Avoid_: Graph drawing, dependency chart

**Blocker**:
An unsolved prerequisite that prevents a Problem from becoming Available in a Roadmap. A Verified prerequisite remains a Blocker until it becomes Solved.
_Avoid_: Dependency, obstacle, missing task

**Dashboard**:
The main Leetgo screen after Onboarding, focused on the user's next best action while showing supporting Profile, gamification, and Roadmap context.
_Avoid_: Home page, stats page, main menu

**Inline Status**:
Short task feedback shown inside the active screen near the related action or section.
_Avoid_: Toast, alert, modal

**Unsupported Size**:
The fallback TUI state shown when the terminal is too small to render Leetgo's primary content clearly. Leetgo's minimum supported TUI size is 60 columns by 18 rows.
_Avoid_: Broken layout, tiny mode, error screen

**Quest Board**:
The RPG skill tree Theme's Dashboard presentation of Next Actions, with the primary recommendation treated as the main quest and secondary actions treated as supporting choices.
_Avoid_: Task board, todo list, menu

**Character HUD**:
The RPG skill tree Theme's presentation of Profile and gamification context, including Level, XP, Streak, and Achievements.
_Avoid_: Stats panel, user card, sidebar

**Next Action**:
A recommended action the user can take from the Dashboard, such as continuing an InProgress Problem, starting an Available Problem, reviewing a blocker, or submitting a solution.
_Avoid_: Task, todo, suggestion

**Review**:
A learning action where the user revisits a previously Solved or difficult Problem to strengthen retention, repair a Weakness, or prepare for a blocked dependent. Review does not block Roadmap completion.
_Avoid_: Redo, repeat, revision

**Review Cycle**:
A bounded Review opportunity created by Leetgo for a specific learning reason, such as Weakness repair, high failed-attempt history, manual Solve validation, or prerequisite refresh. Review XP can be earned at most once per Review Cycle.
_Avoid_: Review session, redo window, practice loop

**Recommendation Reason**:
The user-facing explanation for why a Next Action is recommended, including prerequisite progress, Category reinforcement, unlock impact, or Submission readiness.
_Avoid_: Tooltip, hint, algorithm output

**Roadmap Detail**:
The drill-down view of a Roadmap, focused on its Unlock Path and Stage progress.
_Avoid_: Roadmap page, graph screen

**Roadmap Completion**:
The milestone reached when every Problem in a Roadmap is Solved, shown with completion summary, learning quality, and follow-up recommendations.
_Avoid_: Done state, finished roadmap, end screen

**Stage Detail**:
The drill-down view of one Stage, focused on its Problems, completion, and blockers.
_Avoid_: Category page, module view

**Problem Detail**:
The drill-down view of one Problem, focused on Start, local testing, Submission, and Practice Log history.
_Avoid_: Problem page, challenge view

**Problem Brief**:
Leetgo's concise learning explanation for a Problem, covering the plain-language task, Practice Focus, prerequisite context, unlock impact, and optional starter hints.
_Avoid_: Problem statement, editorial, solution

**Problem Time Estimate**:
An expected practice-session range for a Problem, used to help the user choose work that fits their available time.
_Avoid_: Deadline, timer, exact ETA

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

**Motion Preference**:
The user's choice for animated feedback in Leetgo's CLI and TUI: normal, reduced, or off. Non-interactive command output behaves as motion off.
_Avoid_: Animation setting, effects toggle, motion mode

**Terminal Palette**:
The ANSI-oriented color roles used by a Theme, such as Primary, Success, Warning, Danger, Muted, Border, XP, and Review. Theme identity comes from layout, emphasis, symbols, and motion as much as color.
_Avoid_: Truecolor palette, paint scheme, CSS theme

**Pixel Frame**:
The blocky card and panel treatment used by the RPG skill tree Theme to make Problems, Next Actions, and progress summaries feel like game tiles.
_Avoid_: Rounded card, box skin, pixel art

**Rich Symbols**:
The icon set used by styled TUI Themes when the terminal supports Nerd Fonts or modern emoji rendering. Rich Symbols communicate state and progression rather than acting as decoration.
_Avoid_: Emoji spam, decorative icons, image assets

**Plain Symbols**:
The ASCII-safe fallback symbol set used when Rich Symbols are disabled or unsupported.
_Avoid_: No-icon mode, degraded UI, basic skin

**Workspace**:
The user-owned area where generated Problem files live.
_Avoid_: Project dir, output dir

**Stub**:
The generated source file for a Problem in the user's chosen language.
_Avoid_: Template, skeleton, boilerplate

**TestSuite**:
The generated test file for a Problem.
_Avoid_: Tests, test file, spec

**Problem Manifest**:
A hidden metadata file generated beside a Problem's Stub and TestSuite so Leetgo can identify the Problem when commands are run from inside that Problem's workspace directory. The Problem Manifest file is `.leetgo-problem.toml`.
_Avoid_: Metadata blob, slug file, marker file

**Language**:
A programming language supported for Stub generation, TestSuite generation, and LeetCode Submissions.
_Avoid_: Runtime, template type

### Progress & Gamification

**Attempt**:
A recorded instance of a user working on a Problem: timestamp, duration, pass/fail result from local tests, and optional self-reported difficulty rating.
_Avoid_: Try, submission, run

**Solve Duration**:
The total recorded practice time spent on a Problem before Solve, calculated from the user's Attempts rather than wall-clock time.
_Avoid_: Time elapsed, calendar time, ETA

**Status**:
A Problem's practice state for the user: Locked, Available, InProgress, Verified, or Solved. Locked and Available depend on the selected Roadmap; Verified and Solved carry across Roadmaps.
_Avoid_: State, phase

**Start**:
The act of beginning work on an Available Problem.
_Avoid_: Open, generate, launch

**Solve**:
The act of marking a Problem Solved for the user across all Roadmaps, either through an Accepted LeetCode Submission or a manual user confirmation.
_Avoid_: Finish, complete, close

**Verified**:
A Problem Status meaning the Problem has passed Leetgo's local TestSuite but has not yet received an Accepted LeetCode Submission. Verified Problems earn the local half of XP once and remain candidates for Submission.
_Avoid_: Locally solved, passed, completed

**Manual Solve**:
A user-confirmed Solve without an Accepted LeetCode Submission. Manual Solve unlocks dependent Problems but does not earn XP, and can later be upgraded by an Accepted Submission.
_Avoid_: Skip, force complete, fake solve

**Accepted Solve**:
A Solve backed by an Accepted LeetCode Submission. Accepted Solve earns Submission XP when eligible and is the trusted completion form in Leetgo's interface.
_Avoid_: Official solve, real solve, LeetCode solve

**XP**:
Experience points earned by solving Problems.
_Avoid_: Points, score

**Reward Event**:
A persisted record that XP was awarded for a specific reason, such as local Verify or Accepted Submission. Reward Events make XP awards idempotent.
_Avoid_: XP log, reward grant, points row

**Reward Moment**:
A short user-facing feedback state after meaningful progress, such as Accepted Solve, Manual Solve, Review Cycle completion, Level-up, Achievement, or Roadmap Completion. Reward Moments explain what changed and what the user can do next.
_Avoid_: Animation, toast, celebration screen

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
A Category or Practice Focus where the user's Attempts show repeated failure, high self-reported difficulty, or repeated Review needs.
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

**Practice Log**:
The full chronological learning history for a Problem, including local Attempts, Submission Attempts, Verification, Manual Solve, Accepted Solve, Solve Duration, and optional notes.
_Avoid_: Solve log, post, journal entry

**Session**:
A user's authenticated LeetCode access used for Submissions.
_Avoid_: Login, auth, connection
