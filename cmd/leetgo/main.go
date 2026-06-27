package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/analytics"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/gitexport"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/recommendation"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui"
	"github.com/prettyletto/leetgo/internal/tui/views"
	"github.com/prettyletto/leetgo/internal/workspace"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(args) == 0 {
		return startTUI(cfg)
	}

	switch args[0] {
	case "start":
		return startTUI(cfg)
	case "init":
		return initConfig(cfg)
	case "auth":
		return authenticate()
	case "status":
		return showStatus(cfg)
	case "paths":
		return showPaths()
	case "solve-log":
		return showSolveLog(args[1:])
	case "solve":
		return runManualSolve(cfg, args[1:])
	case "next":
		return runNext(cfg, args[1:])
	case "info":
		return runInfo(cfg, args[1:])
	case "test":
		return runProblemTests(cfg, args[1:])
	case "submit":
		return runProblemSubmit(cfg, args[1:])
	case "roadmap":
		return handleRoadmap(cfg, args[1:])
	case "config":
		return handleConfig(cfg, args[1:])
	case "export":
		return exportData(cfg)
	case "git-export":
		return exportDataToGit(cfg, args[1:])
	case "import":
		return importData(cfg, args[1:])
	case "onboard", "onboarding":
		return runOnboard(cfg, args[1:])
	case "reset":
		return runReset(cfg, args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func startTUI(cfg *config.Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	legacyModel, err := tui.NewModel(cfg, db)
	if err != nil {
		return fmt.Errorf("create TUI: %w", err)
	}

	roadmaps, err := catalog.ListRoadmaps()
	if err != nil {
		return fmt.Errorf("list roadmaps: %w", err)
	}

	languages := languageIDs()

	model, err := tui.NewRootModel(cfg, legacyModel, db, languages, roadmaps)
	if err != nil {
		return fmt.Errorf("create root TUI: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}

	return nil
}

func openStore() (*store.SQLiteStore, error) {
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return nil, err
	}
	return store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
}

func showPaths() error {
	paths, err := config.AppPaths()
	if err != nil {
		return err
	}
	fmt.Printf("Leetgo Paths\n")
	fmt.Printf("  Home:       %s\n", paths.Home)
	fmt.Printf("  Data dir:   %s\n", paths.DataDir)
	fmt.Printf("  Config:     %s\n", paths.ConfigFile)
	fmt.Printf("  Database:   %s\n", paths.Database)
	fmt.Printf("  Session:    %s\n", paths.Session)
	fmt.Printf("  Exports:    %s\n", paths.ExportsDir)
	return nil
}

func validateConfig(cfg *config.Config) error {
	roadmaps, err := roadmapIDs()
	if err != nil {
		return err
	}
	if err := cfg.Validate(languageIDs(), roadmaps); err != nil {
		return err
	}
	return nil
}

func initConfig(cfg *config.Config) error {
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("Config saved to ~/.leetgo/config.toml\n")
	fmt.Printf("Workspace: %s\n", cfg.Workspace)
	fmt.Printf("Roadmap:   %s\n", cfg.Roadmap)
	return nil
}

func showStatus(cfg *config.Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	stats, err := db.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("get stats: %w", err)
	}

	rm, err := catalog.LoadRoadmap(cfg.Roadmap)
	if err != nil {
		return fmt.Errorf("load roadmap: %w", err)
	}
	graph := rm.Graph
	stats.Total = len(graph.Problems)

	fmt.Printf("Leetgo Status\n")
	fmt.Printf("  Roadmap:  %s\n", rm.Title)
	fmt.Printf("  Level:    %d (%d XP)\n", stats.Level, stats.TotalXP)
	fmt.Printf("  Solved:   %d/%d\n", stats.Solved, stats.Total)
	fmt.Printf("  Streak:   %d days (longest: %d)\n", stats.Streak, stats.LongestStreak)

	roadmaps, err := catalog.ListRoadmaps()
	if err == nil && len(roadmaps) > 0 {
		fmt.Printf("  Available Roadmaps:")
		for _, available := range roadmaps {
			marker := ""
			if available.ID == rm.ID {
				marker = "*"
			}
			fmt.Printf(" %s%s", marker, available.ID)
		}
		fmt.Printf("\n")
	}

	achievements, err := db.GetAchievements(ctx)
	if err == nil && len(achievements) > 0 {
		fmt.Printf("  Achievements: %d unlocked\n", len(achievements))
	}

	analyticsEngine := analytics.New(db, graph)
	weaknesses, err := analyticsEngine.DetectWeaknesses(ctx)
	if err == nil && len(weaknesses) > 0 {
		fmt.Printf("  Weaknesses:\n")
		for _, w := range weaknesses {
			fmt.Printf("    - %s (%s)\n", w.Category, w.Reason)
		}
	}

	solved := make(map[int]bool)
	progress, _ := db.GetAllProgress(ctx)
	for id, status := range progress {
		if status == roadmap.StatusSolved {
			solved[id] = true
		}
	}

	if rm.IsComplete(solved) {
		fmt.Println()
		fmt.Println("Roadmap Complete!")
		prov, _ := db.GetSolveProvenanceAll(ctx)
		acceptedCount := 0
		manualCount := 0
		for _, sp := range prov {
			if sp.Kind == "accepted" {
				acceptedCount++
			} else {
				manualCount++
			}
		}
		fmt.Printf("  Solved:     %d/%d\n", stats.Solved, stats.Total)
		fmt.Printf("  Accepted:   %d\n", acceptedCount)
		fmt.Printf("  Manual:     %d\n", manualCount)
		fmt.Printf("  Total XP:   %d\n", stats.TotalXP)

		cycles, _ := db.GetReviewCycles(ctx)
		activeCycles := 0
		for _, rc := range cycles {
			if rc.CompletedAt == nil {
				activeCycles++
			}
		}
		if activeCycles > 0 {
			fmt.Printf("  Active Review Cycles: %d\n", activeCycles)
		}

		if len(rm.NextRoadmaps) > 0 {
			fmt.Printf("  Suggested next: %s\n", rm.NextRoadmaps[0])
		}
	}

	return nil
}

func showSolveLog(args []string) error {
	limit := 10
	if len(args) > 1 {
		return fmt.Errorf("usage: leetgo solve-log [limit]")
	}
	if len(args) == 1 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil || parsed <= 0 {
			return fmt.Errorf("limit must be a positive number")
		}
		limit = parsed
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	logs, err := db.GetSolveLogs(context.Background())
	if err != nil {
		return err
	}
	if len(logs) == 0 {
		fmt.Println("No Solve Logs yet.")
		return nil
	}
	if len(logs) < limit {
		limit = len(logs)
	}

	fmt.Println("Practice Log")
	for _, log := range logs[:limit] {
		when := log.SubmittedAt.Format("2006-01-02 15:04")
		result := fmt.Sprintf("%s (%d/%d)", log.Status, log.PassedTests, log.TotalTests)
		if log.StatusCode == 10 {
			result = fmt.Sprintf("Accepted · %s · %s", log.Runtime, log.Memory)
		}
		fmt.Printf("  %s  #%d %s  %s  %s\n", when, log.ProblemID, log.Slug, log.Language, result)
	}
	return nil
}

func runProblemTests(cfg *config.Config, args []string) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: leetgo test <problem-id|problem-slug|.>")
	}

	problem, language, problemDir, err := resolveProblem(args[0], cfg)
	if err != nil {
		return err
	}

	if _, err := os.Stat(problemDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("problem files not found: start %s first", problem.Slug)
		}
		return fmt.Errorf("stat problem dir: %w", err)
	}
	_, canVerify, _, reason := generator.AutomationSupport(problem)
	if !canVerify {
		return fmt.Errorf("local verification unavailable for #%d %s: %s", problem.ID, problem.Title, reason)
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	progress, _ := db.GetProgress(ctx, problem.ID)

	cmd, err := testCommand(language, problemDir)
	if err != nil {
		return err
	}
	startedAt := time.Now()
	output, testErr := cmd.CombinedOutput()
	testDuration := time.Since(startedAt)
	testOutput := strings.TrimSpace(string(output))
	_ = db.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: problem.ID,
		Timestamp: startedAt,
		Duration:  testDuration,
		Passed:    testErr == nil,
	})

	if testErr != nil {
		statusLabel := "InProgress"
		if progress != nil {
			statusLabel = cliStatusLabel(progress.Status)
		}
		printTestFailed(problem, statusLabel, testOutput)
		return fmt.Errorf("tests failed: %w", testErr)
	}

	alreadyClaimed, err := db.HasRewardEvent(ctx, problem.ID, "verify")
	if err != nil {
		return fmt.Errorf("check verify reward: %w", err)
	}

	if alreadyClaimed {
		currentStatus := "Verified"
		if progress != nil && progress.Status == roadmap.StatusSolved {
			currentStatus = "Solved"
		}
		printTestPassedWithStatus(problem, currentStatus, "already claimed", "run `leetgo submit .`")
		return nil
	}

	xp := store.XPForDifficulty(problem.Difficulty) * 70 / 100
	if xp > 0 {
		if err := db.AddXP(ctx, xp); err != nil {
			return fmt.Errorf("add verify XP: %w", err)
		}
	}

	event := &store.RewardEvent{
		ProblemID: problem.ID,
		Kind:      "verify",
		XP:        xp,
	}
	if err := db.RecordRewardEvent(ctx, event); err != nil {
		return fmt.Errorf("record verify event: %w", err)
	}

	if progress == nil || progress.Status != roadmap.StatusSolved {
		if err := db.SetProgress(ctx, problem.ID, roadmap.StatusVerified); err != nil {
			return fmt.Errorf("update progress: %w", err)
		}
	}

	if err := db.UpdateStreak(ctx); err != nil {
		return fmt.Errorf("update streak: %w", err)
	}

	statusLabel := "Verified"
	if progress != nil && progress.Status == roadmap.StatusSolved {
		statusLabel = "Solved"
	}
	printTestPassedWithStatus(problem, statusLabel, fmt.Sprintf("+%d XP", xp), "run `leetgo submit .` for LeetCode confirmation and bonus XP")
	return nil
}

func runProblemSubmit(cfg *config.Config, args []string) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	skipTests := false
	var problemArg string
	for _, arg := range args {
		switch arg {
		case "--skip-tests":
			skipTests = true
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown submit option: %s", arg)
			}
			if problemArg != "" {
				return fmt.Errorf("usage: leetgo submit [--skip-tests] <problem-id|problem-slug|.>")
			}
			problemArg = arg
		}
	}
	if problemArg == "" {
		return fmt.Errorf("usage: leetgo submit [--skip-tests] <problem-id|problem-slug|.>")
	}

	problem, language, problemDir, err := resolveProblem(problemArg, cfg)
	if err != nil {
		return err
	}

	if _, err := os.Stat(problemDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("problem files not found: start %s first", problem.Slug)
		}
		return fmt.Errorf("stat problem dir: %w", err)
	}
	_, canVerify, canSubmit, reason := generator.AutomationSupport(problem)
	if !canSubmit {
		return fmt.Errorf("submission unavailable for #%d %s: %s", problem.ID, problem.Title, reason)
	}
	if !skipTests && !canVerify {
		return fmt.Errorf("local verification unavailable for #%d %s: %s. Use --skip-tests to submit directly to LeetCode", problem.ID, problem.Title, reason)
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	progress, _ := db.GetProgress(ctx, problem.ID)

	if !skipTests {
		cmd, err := testCommand(language, problemDir)
		if err != nil {
			return err
		}
		startedAt := time.Now()
		output, testErr := cmd.CombinedOutput()
		testDuration := time.Since(startedAt)
		testOutput := strings.TrimSpace(string(output))
		_ = db.RecordAttempt(ctx, &store.AttemptRecord{
			ProblemID: problem.ID,
			Timestamp: startedAt,
			Duration:  testDuration,
			Passed:    testErr == nil,
		})

		if testErr != nil {
			statusLabel := "InProgress"
			if progress != nil {
				statusLabel = cliStatusLabel(progress.Status)
			}
			fmt.Printf("Local tests failed for #%d %s\n", problem.ID, problem.Title)
			fmt.Printf("Status: %s\n", statusLabel)
			if testOutput != "" {
				fmt.Printf("\n%s\n", testOutput)
			}
			fmt.Println("\nFix your solution first, or use --skip-tests to submit anyway.")
			return fmt.Errorf("tests failed: %w", testErr)
		}

		alreadyClaimed, err := db.HasRewardEvent(ctx, problem.ID, "verify")
		if err != nil {
			return fmt.Errorf("check verify reward: %w", err)
		}

		if !alreadyClaimed {
			xp := store.XPForDifficulty(problem.Difficulty) * 70 / 100
			if xp > 0 {
				if err := db.AddXP(ctx, xp); err != nil {
					return fmt.Errorf("add verify XP: %w", err)
				}
			}

			event := &store.RewardEvent{
				ProblemID: problem.ID,
				Kind:      "verify",
				XP:        xp,
			}
			if err := db.RecordRewardEvent(ctx, event); err != nil {
				return fmt.Errorf("record verify event: %w", err)
			}
		}

		if progress == nil || progress.Status != roadmap.StatusSolved {
			if err := db.SetProgress(ctx, problem.ID, roadmap.StatusVerified); err != nil {
				return fmt.Errorf("update progress: %w", err)
			}
		}

		fmt.Printf("Local tests passed for #%d %s\n", problem.ID, problem.Title)
	} else {
		fmt.Printf("Skipping local tests (--skip-tests) for #%d %s\n", problem.ID, problem.Title)
	}

	stubName := strings.ReplaceAll(problem.Slug, "-", "_") + stubExt(language)
	stubPath := filepath.Join(problemDir, stubName)

	code, err := os.ReadFile(stubPath)
	if err != nil {
		return fmt.Errorf("read solution file: %w", err)
	}

	client, err := leetcode.NewClient()
	if err != nil {
		return fmt.Errorf("create leetcode client: %w", err)
	}

	if !client.IsAuthenticated() {
		statusLabel := "InProgress"
		if progress != nil {
			statusLabel = cliStatusLabel(progress.Status)
		}
		printSubmitUnavailable(problem, statusLabel)
		return nil
	}

	submitStartedAt := time.Now()
	result, submitErr := client.Submit(ctx, problem.ID, problem.Slug, leetcodeLang(language), string(code))
	submitDuration := time.Since(submitStartedAt)
	if submitErr != nil {
		return fmt.Errorf("submit failed: %w", submitErr)
	}
	_ = db.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: problem.ID,
		Timestamp: submitStartedAt,
		Duration:  submitDuration,
		Passed:    result.StatusCode == 10,
	})

	solveLog := &store.SolveLogRecord{
		ProblemID:   problem.ID,
		Slug:        problem.Slug,
		Language:    leetcodeLang(language),
		Status:      result.Status,
		StatusCode:  result.StatusCode,
		Runtime:     result.Runtime,
		Memory:      result.Memory,
		TotalTests:  result.TotalTests,
		PassedTests: result.PassedTests,
		Error:       result.Error,
	}

	if err := db.RecordSolveLog(ctx, solveLog); err != nil {
		return fmt.Errorf("record solve log: %w", err)
	}

	if result.StatusCode == 10 {
		alreadyClaimed, err := db.HasRewardEvent(ctx, problem.ID, "submit")
		if err != nil {
			return fmt.Errorf("check submit reward: %w", err)
		}

		if alreadyClaimed {
			printSubmitAccepted(problem, "already claimed", result.Runtime, result.Memory)
		} else {
			xp := store.XPForDifficulty(problem.Difficulty) * 30 / 100
			if xp > 0 {
				if err := db.AddXP(ctx, xp); err != nil {
					return fmt.Errorf("add submit XP: %w", err)
				}
			}

			event := &store.RewardEvent{
				ProblemID: problem.ID,
				Kind:      "submit",
				XP:        xp,
			}
			if err := db.RecordRewardEvent(ctx, event); err != nil {
				return fmt.Errorf("record submit event: %w", err)
			}

			hasVerify, err := db.HasRewardEvent(ctx, problem.ID, "verify")
			if err != nil {
				return fmt.Errorf("check verify reward: %w", err)
			}
			totalXP := xp
			if !hasVerify {
				verifyXP := store.XPForDifficulty(problem.Difficulty) * 70 / 100
				if verifyXP > 0 {
					if err := db.AddXP(ctx, verifyXP); err != nil {
						return fmt.Errorf("add verify XP: %w", err)
					}
				}
				verifyEvent := &store.RewardEvent{
					ProblemID: problem.ID,
					Kind:      "verify",
					XP:        verifyXP,
				}
				if err := db.RecordRewardEvent(ctx, verifyEvent); err != nil {
					return fmt.Errorf("record verify event: %w", err)
				}
				totalXP += verifyXP
			}
			printSubmitAccepted(problem, fmt.Sprintf("+%d XP", totalXP), result.Runtime, result.Memory)
		}

		if err := db.SetProgress(ctx, problem.ID, roadmap.StatusSolved); err != nil {
			return fmt.Errorf("update progress: %w", err)
		}

		recordAcceptedProvenanceCLI(ctx, db, problem.ID)

		if err := db.UpdateStreak(ctx); err != nil {
			return fmt.Errorf("update streak: %w", err)
		}
	} else {
		statusLabel := "InProgress"
		if progress != nil {
			statusLabel = cliStatusLabel(progress.Status)
		}
		printSubmitRejected(problem, statusLabel, result)
	}

	return nil
}

func recordAcceptedProvenanceCLI(ctx context.Context, db store.Store, problemID int) {
	sp, err := db.GetSolveProvenance(ctx, problemID)
	if err != nil {
		return
	}

	logs, err := db.GetSolveLogsForProblem(ctx, problemID)
	var logID *int
	if err == nil && len(logs) > 0 {
		latestAccepted := -1
		for _, l := range logs {
			if l.StatusCode == 10 && l.ID > latestAccepted {
				latestAccepted = l.ID
			}
		}
		if latestAccepted > 0 {
			logID = &latestAccepted
		}
	}

	if sp != nil && sp.Kind == "manual" {
		_ = db.RecordSolveProvenance(ctx, &store.SolveProvenance{
			ProblemID:  problemID,
			Kind:       "accepted",
			Note:       sp.Note,
			SolveLogID: logID,
		})
	} else {
		_ = db.RecordSolveProvenance(ctx, &store.SolveProvenance{
			ProblemID:  problemID,
			Kind:       "accepted",
			SolveLogID: logID,
		})
	}
}

func runManualSolve(cfg *config.Config, args []string) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	manual := false
	yes := false
	note := ""
	var problemArg string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--manual":
			manual = true
		case "--yes", "-y":
			yes = true
		case "--note":
			i++
			if i >= len(args) {
				return fmt.Errorf("usage: leetgo solve --manual [--yes] [--note <text>] <problem-id|problem-slug|.>")
			}
			note = args[i]
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown solve option: %s", arg)
			}
			if problemArg != "" {
				return fmt.Errorf("usage: leetgo solve --manual [--yes] [--note <text>] <problem-id|problem-slug|.>")
			}
			problemArg = arg
		}
	}

	if !manual {
		return fmt.Errorf("usage: leetgo solve --manual [--yes] [--note <text>] <problem-id|problem-slug|.>")
	}
	if problemArg == "" {
		return fmt.Errorf("usage: leetgo solve --manual [--yes] [--note <text>] <problem-id|problem-slug|.>")
	}

	problem, _, _, err := resolveProblem(problemArg, cfg)
	if err != nil {
		return err
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	progress, err := db.GetProgress(ctx, problem.ID)
	if err != nil {
		return fmt.Errorf("get progress: %w", err)
	}
	if progress != nil && progress.Status == roadmap.StatusSolved {
		fmt.Printf("Already Solved: #%d %s\n", problem.ID, problem.Title)
		return nil
	}

	if !yes {
		fmt.Printf("Mark as manually solved? This will unlock dependent Problems, but you will not earn XP unless LeetCode accepts a Submission later.\n")
		fmt.Printf("Problem: #%d %s\n", problem.ID, problem.Title)
		fmt.Print("Proceed? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := db.SetProgress(ctx, problem.ID, roadmap.StatusSolved); err != nil {
		return fmt.Errorf("update progress: %w", err)
	}

	if err := db.RecordSolveProvenance(ctx, &store.SolveProvenance{
		ProblemID: problem.ID,
		Kind:      "manual",
		Note:      note,
	}); err != nil {
		return fmt.Errorf("record solve provenance: %w", err)
	}

	if err := db.UpdateStreak(ctx); err != nil {
		return fmt.Errorf("update streak: %w", err)
	}

	fmt.Printf("Manually Solved: #%d %s\n", problem.ID, problem.Title)
	fmt.Println("Reward: none (Manual Solve earns no XP)")
	if note != "" {
		fmt.Printf("Note: %s\n", note)
	}

	return nil
}

func runNext(cfg *config.Config, args []string) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	showAll := false
	doStart := false
	for _, arg := range args {
		switch arg {
		case "--all":
			showAll = true
		case "--start":
			doStart = true
		default:
			return fmt.Errorf("unknown next option: %s", arg)
		}
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	rm, err := catalog.LoadRoadmap(cfg.Roadmap)
	if err != nil {
		return fmt.Errorf("load roadmap: %w", err)
	}

	ctx := context.Background()
	calc := recommendation.NewCalculator(db, rm)
	actions, err := calc.Calculate(ctx)
	if err != nil {
		return fmt.Errorf("calculate next actions: %w", err)
	}

	if len(actions) == 0 {
		fmt.Println("No Next Actions available.")
		return nil
	}

	primary := actions[0]

	if showAll {
		fmt.Println("Next Actions:")
		for i, a := range actions {
			label := formatNextActionLabel(a.Kind)
			fmt.Printf("  %d. %s  %s\n", i+1, label, a.Title)
			fmt.Printf("     %s\n", a.Reason)
			if a.PracticeFocus != "" {
				fmt.Printf("     Focus: %s\n", a.PracticeFocus)
			}
		}
		return nil
	}

	if doStart {
		if primary.Kind != recommendation.KindStart {
			fmt.Printf("Primary action is %s, not Start.\n", primary.Kind)
			fmt.Printf("Run: leetgo %s\n", actionToCommand(primary.Kind))
			return nil
		}
		return doStartProblem(cfg, primary.ProblemID)
	}

	label := formatNextActionLabel(primary.Kind)
	fmt.Printf("Next: %s\n", label)
	fmt.Printf("  %s\n", primary.Title)
	fmt.Printf("  %s\n", primary.Reason)
	if primary.PracticeFocus != "" {
		fmt.Printf("  Practice Focus: %s\n", primary.PracticeFocus)
	}
	fmt.Printf("\n  Run leetgo next --all to see all options.\n")
	return nil
}

func formatNextActionLabel(kind recommendation.ActionKind) string {
	switch kind {
	case "manual_solve":
		return "Manual Solve"
	case "connect_leetcode":
		return "Connect LeetCode"
	case "view_roadmap_completion":
		return "Roadmap Completion"
	default:
		s := string(kind)
		return strings.ToUpper(s[:1]) + s[1:]
	}
}

func actionToCommand(kind recommendation.ActionKind) string {
	switch kind {
	case recommendation.KindSubmit:
		return "submit <problem>"
	case recommendation.KindManualSolve:
		return "solve --manual <problem>"
	case recommendation.KindContinue:
		return "start (then test/submit the InProgress problem)"
	case recommendation.KindReview:
		return "start (then test the Review problem)"
	case recommendation.KindConnectLeetCode:
		return "auth"
	default:
		return string(kind)
	}
}

func doStartProblem(cfg *config.Config, problemID int) error {
	rm, err := catalog.LoadRoadmap(cfg.Roadmap)
	if err != nil {
		return fmt.Errorf("load roadmap: %w", err)
	}

	problem, ok := rm.Graph.Problems[problemID]
	if !ok {
		return fmt.Errorf("problem %d not found in roadmap %q", problemID, cfg.Roadmap)
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	manager := workspace.New(cfg.Workspace, generator.New())
	problemDir := manager.ProblemDir(problem)

	if err := workspace.EnsureManifestWritable(problemDir, problem.ID); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	if err := db.SetProgress(ctx, problem.ID, roadmap.StatusInProgress); err != nil {
		return fmt.Errorf("update progress: %w", err)
	}

	stubPath, testPath, err := manager.Generate(problem, generator.Language(cfg.Language))
	if err != nil {
		return fmt.Errorf("generate files: %w", err)
	}

	stage := problem.Stage
	if stage == "" {
		stage = string(problem.Category)
	}
	m := &workspace.Manifest{
		ProblemID:     problem.ID,
		Slug:          problem.Slug,
		Roadmap:       cfg.Roadmap,
		Stage:         stage,
		Language:      cfg.Language,
		StubPath:      filepath.Base(stubPath),
		TestsuitePath: filepath.Base(testPath),
	}
	if err := workspace.WriteManifest(problemDir, m); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	fmt.Printf("Started: #%d %s\n", problem.ID, problem.Title)
	fmt.Printf("Stub: %s\n", stubPath)
	fmt.Printf("Test: %s\n", testPath)

	editor := cfg.Editor
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor != "" {
		parts := strings.Fields(editor)
		args := append(parts[1:], stubPath, testPath)
		cmd := exec.Command(parts[0], args...)
		cmd.Dir = problemDir
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("open editor: %w", err)
		}
	}
	return nil
}

func runInfo(cfg *config.Config, args []string) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	var problem *roadmap.Problem
	var rm *roadmap.Roadmap
	var db store.Store
	var err error

	db, err = openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if len(args) == 0 {
		searchDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		m, _, err := workspace.ReadManifest(searchDir)
		if err != nil {
			return fmt.Errorf("read manifest from %q: %w (did you mean `leetgo info <id-or-slug>`?)", searchDir, err)
		}
		if m == nil {
			return fmt.Errorf("not in a problem workspace and no problem specified. Usage: leetgo info <id-or-slug>")
		}
		rm, err = catalog.LoadRoadmap(m.Roadmap)
		if err != nil {
			return fmt.Errorf("load roadmap %q: %w", m.Roadmap, err)
		}
		p, ok := rm.Graph.Problems[m.ProblemID]
		if !ok {
			return fmt.Errorf("problem %d not found in roadmap %q", m.ProblemID, m.Roadmap)
		}
		problem = p
	} else {
		arg := args[0]
		rm, err = catalog.LoadRoadmap(cfg.Roadmap)
		if err != nil {
			return fmt.Errorf("load roadmap: %w", err)
		}
		p, err := findProblem(rm.Graph, arg)
		if err != nil {
			return fmt.Errorf("could not resolve %q: not a valid problem ID or slug", arg)
		}
		problem = p
	}

	ctx := context.Background()
	progress, _ := db.GetProgress(ctx, problem.ID)
	status := roadmap.StatusAvailable
	if progress != nil {
		status = progress.Status
	} else {
		allProgress, _ := db.GetAllProgress(ctx)
		locked := false
		for _, prereq := range problem.Prerequisites {
			if allProgress[prereq] != roadmap.StatusSolved {
				locked = true
				break
			}
		}
		if locked {
			status = roadmap.StatusLocked
		}
	}

	printProblemInfo(problem, status, db, rm)
	return nil
}

func printProblemInfo(p *roadmap.Problem, status roadmap.Status, db store.Store, rm *roadmap.Roadmap) {
	fmt.Printf("#%d %s\n", p.ID, p.Title)
	fmt.Printf("  Difficulty: %s\n", p.Difficulty)
	fmt.Printf("  Category:  %s\n", p.Category)
	fmt.Printf("  Status:    %s\n", statusLabel(status))

	sp, _ := db.GetSolveProvenance(context.Background(), p.ID)
	if sp != nil && status == roadmap.StatusSolved {
		kindLabel := "Accepted Solve"
		if sp.Kind == "manual" {
			kindLabel = "Manual Solve"
		}
		fmt.Printf("  Solved via: %s\n", kindLabel)
		if sp.Note != "" {
			fmt.Printf("  Note:       %s\n", sp.Note)
		}
	}

	fmt.Println()

	if p.Summary != "" {
		fmt.Println("Problem Brief")
		fmt.Printf("  Summary:        %s\n", p.Summary)
		if p.PracticeFocus != "" {
			fmt.Printf("  Practice Focus: %s\n", p.PracticeFocus)
		}
		if p.WhyNow != "" {
			fmt.Printf("  Why now:        %s\n", p.WhyNow)
		}
		if p.UnlockImpact != "" {
			fmt.Printf("  Unlock Impact:  %s\n", p.UnlockImpact)
		}
		if p.ProblemTimeEstimate != "" {
			fmt.Printf("  Time Estimate:  %s\n", p.ProblemTimeEstimate)
		}
		fmt.Println()
	}

	fmt.Println("Progression")
	if len(p.Prerequisites) == 0 {
		fmt.Println("  Prerequisites: none")
	} else {
		fmt.Println("  Prerequisites:")
		for _, pid := range p.Prerequisites {
			if pp, ok := rm.Graph.Problems[pid]; ok {
				fmt.Printf("    #%d %s\n", pp.ID, pp.Title)
			} else {
				fmt.Printf("    #%d\n", pid)
			}
		}
	}

	if status == roadmap.StatusLocked {
		fmt.Println("  Blockers:")
		allProgress, _ := db.GetAllProgress(context.Background())
		for _, pid := range p.Prerequisites {
			if allProgress[pid] != roadmap.StatusSolved {
				if pp, ok := rm.Graph.Problems[pid]; ok {
					fmt.Printf("    #%d %s\n", pp.ID, pp.Title)
				} else {
					fmt.Printf("    #%d\n", pid)
				}
			}
		}
	}

	unlocks := c.infoUnlocks(p.ID, rm.Graph, db)
	if len(unlocks) > 0 {
		fmt.Println("  Unlocks:")
		for _, u := range unlocks {
			fmt.Printf("    #%d %s (%s)\n", u.ID, u.Title, u.Difficulty)
		}
	}

	if len(unlocks) > 0 {
		indirect := c.infoIndirectUnlocks(p.ID, rm.Graph, db, 2)
		if len(indirect) > 0 {
			fmt.Println("  Builds toward:")
			for _, u := range indirect {
				fmt.Printf("    #%d %s\n", u.ID, u.Title)
			}
		}
	}
	fmt.Println()

	entries := tui.BuildPracticeLog(db, p.ID)
	if len(entries) > 0 {
		fmt.Println("Practice")
		for _, e := range entries {
			fmt.Printf("  %s  %s", e.Timestamp.Format("2006-01-02 15:04"), e.Kind)
			if e.Detail != "" {
				fmt.Printf(" · %s", e.Detail)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Println("Actions")
	if status == roadmap.StatusLocked {
		fmt.Println("  Solve prerequisites first, then start the problem.")
	} else if status == roadmap.StatusAvailable {
		fmt.Println("  leetgo next --start    (to begin working)")
	} else if status == roadmap.StatusInProgress {
		fmt.Println("  leetgo test .          (run local tests)")
		fmt.Println("  leetgo submit .        (submit to LeetCode)")
	} else if status == roadmap.StatusVerified {
		fmt.Println("  leetgo submit .        (submit to LeetCode)")
		fmt.Println("  leetgo solve --manual . (mark manually solved)")
	} else if status == roadmap.StatusSolved {
		fmt.Println("  leetgo test .          (practice again)")
	}
}

type infoHelper struct{}

var c infoHelper

func (c infoHelper) infoUnlocks(problemID int, graph *roadmap.Graph, db store.Store) []*roadmap.Problem {
	var result []*roadmap.Problem
	progress, _ := db.GetAllProgress(context.Background())
	for _, p := range graph.Problems {
		for _, prereq := range p.Prerequisites {
			if prereq == problemID {
				if progress[p.ID] != roadmap.StatusSolved {
					result = append(result, p)
				}
			}
		}
	}
	return result
}

func (c infoHelper) infoIndirectUnlocks(problemID int, graph *roadmap.Graph, db store.Store, depth int) []*roadmap.Problem {
	if depth <= 0 {
		return nil
	}
	var result []*roadmap.Problem
	progress, _ := db.GetAllProgress(context.Background())
	seen := make(map[int]bool)
	for _, p := range graph.Problems {
		for _, prereq := range p.Prerequisites {
			if prereq == problemID {
				if !seen[p.ID] && progress[p.ID] != roadmap.StatusSolved {
					seen[p.ID] = true
					result = append(result, p)
					sub := c.infoIndirectUnlocks(p.ID, graph, db, depth-1)
					for _, s := range sub {
						if !seen[s.ID] {
							seen[s.ID] = true
							result = append(result, s)
						}
					}
				}
			}
		}
	}
	return result
}

func statusLabel(status roadmap.Status) string {
	switch status {
	case roadmap.StatusLocked:
		return "Locked"
	case roadmap.StatusAvailable:
		return "Available"
	case roadmap.StatusInProgress:
		return "In Progress"
	case roadmap.StatusVerified:
		return "Verified"
	case roadmap.StatusSolved:
		return "Solved"
	default:
		return string(status)
	}
}

func printTestPassed(problem *roadmap.Problem, reward, next string) {
	printTestPassedWithStatus(problem, "Verified", reward, next)
}

func printTestPassedWithStatus(problem *roadmap.Problem, status, reward, next string) {
	fmt.Printf("Leetgo TestSuite passed for #%d %s\n", problem.ID, problem.Title)
	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Reward: %s\n", reward)
	fmt.Printf("Next: %s\n", next)
	fmt.Println()
	fmt.Println(renderCLIRewardMoment(views.RewardMoment{
		Title:   "Problem Verified",
		Subject: fmt.Sprintf("#%d %s", problem.ID, problem.Title),
		Reward:  reward,
		Next:    next,
		Actions: []string{"leetgo submit .", "leetgo next"},
	}))
}

func printTestFailed(problem *roadmap.Problem, status, output string) {
	fmt.Printf("Leetgo TestSuite failed for #%d %s\n", problem.ID, problem.Title)
	fmt.Printf("Status: %s\n", status)
	if output != "" {
		fmt.Printf("\n%s\n", output)
	}
}

func printSubmitAccepted(problem *roadmap.Problem, reward, runtime, memory string) {
	fmt.Printf("LeetCode Accepted for #%d %s\n", problem.ID, problem.Title)
	fmt.Println("Status: Solved")
	fmt.Printf("Reward: %s\n", reward)
	if runtime != "" {
		fmt.Printf("Runtime: %s\n", runtime)
	}
	if memory != "" {
		fmt.Printf("Memory: %s\n", memory)
	}
	highlights := make([]string, 0, 2)
	if runtime != "" {
		highlights = append(highlights, "Runtime: "+runtime)
	}
	if memory != "" {
		highlights = append(highlights, "Memory: "+memory)
	}
	fmt.Println()
	fmt.Println(renderCLIRewardMoment(views.RewardMoment{
		Title:                "Problem Solved",
		Subject:              fmt.Sprintf("#%d %s", problem.ID, problem.Title),
		Reward:               reward,
		AdditionalHighlights: highlights,
		Actions:              []string{"leetgo next", "leetgo info " + problem.Slug},
	}))
}

func renderCLIRewardMoment(moment views.RewardMoment) string {
	if isStdoutTerminal() {
		return views.RenderCLIRewardMoment(moment)
	}
	moment.AdditionalHighlights = append([]string{"Output: static non-TTY"}, moment.AdditionalHighlights...)
	return views.RenderCLIRewardMoment(moment)
}

func isStdoutTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func printSubmitUnavailable(problem *roadmap.Problem, status string) {
	fmt.Printf("Submission unavailable for #%d %s\n", problem.ID, problem.Title)
	fmt.Printf("Status: %s\n", status)
	fmt.Println("Reward: none")
	fmt.Println("Error: Session expired. Run `leetgo auth` to reconnect.")
}

func printSubmitRejected(problem *roadmap.Problem, status string, result *leetcode.SubmissionResult) {
	fmt.Printf("LeetCode rejected #%d %s\n", problem.ID, problem.Title)
	fmt.Printf("Status: %s\n", status)
	fmt.Printf("%s (%d/%d tests passed)\n", result.Status, result.PassedTests, result.TotalTests)
	if result.Error != "" {
		fmt.Printf("\n%s\n", result.Error)
	}
}

func cliStatusLabel(status roadmap.Status) string {
	s := string(status)
	if s == "" {
		return "InProgress"
	}
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func stubExt(lang string) string {
	switch lang {
	case "go":
		return ".go"
	case "python":
		return ".py"
	case "typescript":
		return ".ts"
	case "java":
		return ".java"
	case "cpp":
		return ".cpp"
	case "javascript":
		return ".js"
	case "rust":
		return ".rs"
	case "csharp":
		return ".cs"
	default:
		return ".go"
	}
}

func leetcodeLang(lang string) string {
	switch lang {
	case "go":
		return "golang"
	case "python":
		return "python3"
	case "typescript":
		return "typescript"
	case "java":
		return "java"
	case "cpp":
		return "cpp"
	case "javascript":
		return "javascript"
	case "rust":
		return "rust"
	case "csharp":
		return "csharp"
	default:
		return lang
	}
}

func findProblem(graph *roadmap.Graph, value string) (*roadmap.Problem, error) {
	if id, err := strconv.Atoi(value); err == nil {
		if problem, ok := graph.Problems[id]; ok {
			return problem, nil
		}
		return nil, fmt.Errorf("problem id not found in roadmap: %d", id)
	}
	for _, problem := range graph.Problems {
		if problem.Slug == value {
			return problem, nil
		}
	}
	return nil, fmt.Errorf("problem slug not found in roadmap: %s", value)
}

func resolveSearchDir(arg string) (string, error) {
	if arg == "." {
		return os.Getwd()
	}
	if strings.Contains(arg, "/") || strings.Contains(arg, string(os.PathSeparator)) {
		return arg, nil
	}
	info, err := os.Stat(arg)
	if err == nil && info.IsDir() {
		return arg, nil
	}
	return "", nil
}

func resolveProblem(arg string, cfg *config.Config) (*roadmap.Problem, string, string, error) {
	searchDir, err := resolveSearchDir(arg)
	if err != nil {
		return nil, "", "", err
	}

	if searchDir != "" {
		m, dir, err := workspace.ReadManifest(searchDir)
		if err != nil {
			return nil, "", "", fmt.Errorf("read manifest: %w", err)
		}
		if m != nil {
			rm, err := catalog.LoadRoadmap(m.Roadmap)
			if err != nil {
				return nil, "", "", fmt.Errorf("load roadmap %q from manifest: %w", m.Roadmap, err)
			}
			p, ok := rm.Graph.Problems[m.ProblemID]
			if !ok {
				return nil, "", "", fmt.Errorf("problem %d not found in roadmap %q", m.ProblemID, m.Roadmap)
			}
			return p, m.Language, dir, nil
		}
	}

	rm, err := catalog.LoadRoadmap(cfg.Roadmap)
	if err != nil {
		return nil, "", "", fmt.Errorf("load roadmap: %w", err)
	}
	p, err := findProblem(rm.Graph, arg)
	if err != nil {
		return nil, "", "", fmt.Errorf("could not resolve %q: no manifest found and not a valid problem ID or slug", arg)
	}

	manager := workspace.New(cfg.Workspace, generator.New())
	problemDir := manager.ProblemDir(p)
	return p, cfg.Language, problemDir, nil
}

func testCommand(language, dir string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	switch language {
	case "go":
		cmd = exec.Command("go", "test", ".")
	case "python":
		cmd = exec.Command("python", "-m", "pytest")
	case "typescript":
		cmd = exec.Command("npm", "test")
	case "javascript":
		cmd = exec.Command("npm", "test")
	case "java":
		cmd = exec.Command("mvn", "test")
	case "cpp":
		cmd = exec.Command("sh", "-c", "g++ -std=c++17 *.cpp -o /tmp/leetgo-cpp-test && /tmp/leetgo-cpp-test")
	case "rust":
		cmd = exec.Command("sh", "-c", "rustc --test *_test.rs -o /tmp/leetgo-rust-test && /tmp/leetgo-rust-test")
	case "csharp":
		cmd = exec.Command("dotnet", "test")
	default:
		return nil, fmt.Errorf("unsupported language %q", language)
	}
	cmd.Dir = dir
	return cmd, nil
}

func handleRoadmap(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: leetgo roadmap <list|use>")
	}

	switch args[0] {
	case "list":
		return listRoadmaps(cfg)
	case "use":
		if len(args) != 2 {
			return fmt.Errorf("usage: leetgo roadmap use <roadmap-id>")
		}
		return setConfigValue(cfg, "roadmap", args[1])
	default:
		return fmt.Errorf("unknown roadmap command: %s", args[0])
	}
}

func listRoadmaps(cfg *config.Config) error {
	roadmaps, err := catalog.ListRoadmaps()
	if err != nil {
		return err
	}
	for _, rm := range roadmaps {
		marker := " "
		if rm.ID == cfg.Roadmap {
			marker = "*"
		}
		fmt.Printf("%s %s\t%s\n", marker, rm.ID, rm.Title)
	}
	return nil
}

func handleConfig(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		fmt.Printf("onboarding_complete = %v\n", cfg.OnboardingComplete)
		fmt.Printf("display_name = %s\n", cfg.DisplayName)
		fmt.Printf("workspace = %s\n", cfg.Workspace)
		fmt.Printf("editor = %s\n", cfg.Editor)
		fmt.Printf("language = %s\n", cfg.Language)
		fmt.Printf("roadmap = %s\n", cfg.Roadmap)
		fmt.Printf("theme = %s\n", cfg.Theme)
		fmt.Printf("symbol_mode = %s\n", cfg.SymbolMode)
		fmt.Printf("motion_preference = %s\n", cfg.MotionPreference)
		fmt.Printf("git_export_enabled = %v\n", cfg.GitExportEnabled)
		fmt.Printf("git_export_repo = %s\n", cfg.GitExportRepo)
		return nil
	}
	if len(args) != 3 || args[0] != "set" {
		return fmt.Errorf("usage: leetgo config set <workspace|editor|language|roadmap|display-name|theme|symbol-mode|motion|git-export-enabled|git-export-repo> <value>")
	}
	return setConfigValue(cfg, args[1], args[2])
}

func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "workspace":
		cfg.Workspace = value
	case "editor":
		cfg.Editor = value
	case "language":
		if !slices.Contains(languageIDs(), value) {
			return fmt.Errorf("unsupported language %q", value)
		}
		cfg.Language = value
	case "roadmap":
		roadmaps, err := roadmapIDs()
		if err != nil {
			return err
		}
		if !slices.Contains(roadmaps, value) {
			return fmt.Errorf("unknown roadmap %q", value)
		}
		cfg.Roadmap = value
	case "display-name":
		cfg.DisplayName = strings.TrimSpace(value)
	case "theme":
		cfg.Theme = value
		cfg.ApplyDefaults()
		if cfg.Theme != "adaptive" {
			return fmt.Errorf("unknown theme %q", value)
		}
	case "symbol-mode":
		if !slices.Contains(config.ValidSymbolModes, value) {
			return fmt.Errorf("unknown symbol_mode %q", value)
		}
		cfg.SymbolMode = value
	case "motion":
		if !slices.Contains(config.ValidMotionPreferences, value) {
			return fmt.Errorf("unknown motion_preference %q", value)
		}
		cfg.MotionPreference = value
	case "git-export-enabled":
		switch value {
		case "true", "1", "yes":
			cfg.GitExportEnabled = true
		case "false", "0", "no":
			cfg.GitExportEnabled = false
		default:
			return fmt.Errorf("git-export-enabled must be true or false, got %q", value)
		}
	case "git-export-repo":
		cfg.GitExportRepo = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := validateConfig(cfg); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("Updated %s = %s\n", key, value)
	return nil
}

func languageIDs() []string {
	languages := generator.New().Languages()
	ids := make([]string, len(languages))
	for i, lang := range languages {
		ids[i] = string(lang)
	}
	return ids
}

func roadmapIDs() ([]string, error) {
	roadmaps, err := catalog.ListRoadmaps()
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(roadmaps))
	for i, rm := range roadmaps {
		ids[i] = rm.ID
	}
	return ids, nil
}

func authenticate() error {
	client, err := leetcode.NewClient()
	if err != nil {
		return err
	}

	if client.IsAuthenticated() {
		fmt.Println("Already authenticated with LeetCode")
		return nil
	}

	fmt.Println("Starting LeetCode authentication...")
	if err := client.Authenticate(context.Background()); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}

func exportData(cfg *config.Config) error {
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return err
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	exportDir := filepath.Join(dataDir, "exports")
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return fmt.Errorf("create exports dir: %w", err)
	}

	filename := fmt.Sprintf("leetgo-export-%s.json", time.Now().Format("2006-01-02"))
	exportPath := filepath.Join(exportDir, filename)

	if err := store.ExportToFile(context.Background(), db, exportPath); err != nil {
		return fmt.Errorf("export: %w", err)
	}

	fmt.Printf("Exported to %s\n", exportPath)
	return nil
}

func exportDataToGit(cfg *config.Config, args []string) error {
	options, err := parseGitExportArgs(args)
	if err != nil {
		return err
	}
	repoDir := options.repoDir
	info, err := os.Stat(repoDir)
	if err != nil {
		return fmt.Errorf("stat repo dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("repo dir is not a directory: %s", repoDir)
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		return fmt.Errorf("repo dir is not a git repository: %s", repoDir)
	}

	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return err
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	result, err := gitexport.ExportWithOptions(context.Background(), repoDir, dataDir, db, gitexport.ExportOptions{Commit: options.commit})
	if err != nil {
		return fmt.Errorf("git export: %w", err)
	}

	fmt.Printf("Git export written to %s\n", result.Path)
	fmt.Printf("Export Identity: %s (%s)\n", result.Identity.ID, result.Identity.Source)
	if options.commit {
		if result.NoChanges {
			fmt.Println("No export changes to commit.")
		} else {
			fmt.Printf("Committed export: %s\n", result.CommitHash)
		}
	}
	return nil
}

type gitExportArgs struct {
	repoDir string
	commit  bool
}

func parseGitExportArgs(args []string) (*gitExportArgs, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("usage: leetgo git-export <repo-dir> [--commit]")
	}
	parsed := &gitExportArgs{}
	for _, arg := range args {
		switch arg {
		case "--commit":
			parsed.commit = true
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown git-export option: %s", arg)
			}
			if parsed.repoDir != "" {
				return nil, fmt.Errorf("usage: leetgo git-export <repo-dir> [--commit]")
			}
			parsed.repoDir = arg
		}
	}
	if parsed.repoDir == "" {
		return nil, fmt.Errorf("usage: leetgo git-export <repo-dir> [--commit]")
	}
	return parsed, nil
}

func importData(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: leetgo import <file.json>")
	}

	db, err := openStore()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := store.ImportFromFile(context.Background(), db, args[0]); err != nil {
		return fmt.Errorf("import: %w", err)
	}

	fmt.Printf("Imported from %s\n", args[0])
	return nil
}

func runOnboard(cfg *config.Config, args []string) error {
	fresh := false
	for _, arg := range args {
		switch arg {
		case "--fresh":
			fresh = true
		default:
			return fmt.Errorf("unknown onboard option: %s", arg)
		}
	}

	if fresh {
		workspace, err := config.DefaultWorkspace()
		if err != nil {
			return err
		}
		cfg.DisplayName = ""
		cfg.Language = "go"
		cfg.Workspace = workspace
		cfg.Roadmap = "from-zero-to-hero"
		cfg.Theme = "rpg-skill-tree"
		cfg.SymbolMode = "rich"
		cfg.MotionPreference = "normal"
		cfg.GitExportEnabled = false
		cfg.GitExportRepo = ""
	}

	cfg.OnboardingComplete = false
	cfg.OnboardingVersion = 0

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	if fresh {
		fmt.Println("Onboarding reset to fresh defaults. Run `leetgo start` to begin again.")
	} else {
		fmt.Println("Onboarding will run on next start. Run `leetgo start` to open it.")
	}
	fmt.Println("Your progress, XP, Solve Logs, Streaks, and Achievements are preserved.")
	return nil
}

func runReset(cfg *config.Config, args []string) error {
	force := false
	for _, arg := range args {
		switch arg {
		case "--force", "-f":
			force = true
		default:
			return fmt.Errorf("unknown reset option: %s (use --force to confirm)", arg)
		}
	}

	if !force {
		return fmt.Errorf("this will delete all progress, XP, and workspace files.\nRun `leetgo reset --force` to confirm.")
	}

	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return err
	}

	// Delete database
	dbPath := filepath.Join(dataDir, "leetgo.db")
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove database: %w", err)
	}
	fmt.Printf("Deleted %s\n", dbPath)

	// Delete workspace
	workspace := cfg.Workspace
	if workspace != "" {
		if err := os.RemoveAll(workspace); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove workspace: %w", err)
		}
		fmt.Printf("Deleted %s\n", workspace)
	}

	// Reset config to fresh defaults (same as onboard --fresh)
	defaultWorkspace, err := config.DefaultWorkspace()
	if err != nil {
		return err
	}
	cfg.DisplayName = ""
	cfg.Language = "go"
	cfg.Workspace = defaultWorkspace
	cfg.Roadmap = "from-zero-to-hero"
	cfg.Theme = "rpg-skill-tree"
	cfg.SymbolMode = "rich"
	cfg.MotionPreference = "normal"
	cfg.GitExportEnabled = false
	cfg.GitExportRepo = ""
	cfg.OnboardingComplete = false
	cfg.OnboardingVersion = 0

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println("Config reset to fresh defaults.")
	fmt.Println("Everything wiped. Run `leetgo start` to begin onboarding from scratch.")
	return nil
}
