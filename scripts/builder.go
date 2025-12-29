package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"mockserver/pkg/appinfo"
)

// --- CONFIGURATION ---
const (
	AppName = "mockserver"
	NpmDir  = "npm/bin"
	MainPkg = "."
)

// --- STYLING & COLORS ---
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
	White  = "\033[97m"

	IconCheck = "✔"
	IconX     = "✖"
	IconArrow = "➜"
)

// --- TARGETS ---
type Target struct {
	OS       string
	Arch     string
	FileName string
}

var allTargets = []Target{
	{"linux", "amd64", "mockserver-linux"},
	{"linux", "arm64", "mockserver-linux-arm64"},
	{"darwin", "amd64", "mockserver-macos"},
	{"darwin", "arm64", "mockserver-macos-arm64"},
	{"windows", "amd64", "mockserver.exe"},
}

// BuildResult structure for Pterm Table
type BuildResult struct {
	Platform string
	Status   string
	Duration time.Duration
	Artifact string
}

var results []BuildResult
var reader = bufio.NewReader(os.Stdin)

// Global Config Variables
var (
	binDir    string
	buildAll  bool
	copyNpm   bool
	isCurrent bool
)

func main() {
	flag.StringVar(&binDir, "out", "bin", "Output directory")

	flag.BoolVar(&buildAll, "all", false, "Target: Build for ALL platforms")
	flag.BoolVar(&isCurrent, "current", false, "Target: Build for CURRENT platform only")

	flag.BoolVar(&copyNpm, "npm", false, "Action: Copy artifacts to npm/bin directory")

	flag.Parse()

	// Default Behavior: if not choose select -current flag
	if !buildAll && !isCurrent {
		isCurrent = true
	}

	startTotal := time.Now()
	printBanner()

	// Mode Display
	targetMode := "CURRENT OS"
	if buildAll {
		targetMode = "ALL PLATFORMS"
	}

	actionMode := "BUILD ONLY"
	if copyNpm {
		actionMode = "BUILD + NPM DISTRIBUTION"
	}

	version := appinfo.Version
	buildDate := time.Now().Format(time.RFC3339)

	printSection("INITIALIZATION")
	fmt.Printf("  %sTargets    :%s %s\n", Gray, White, targetMode)
	fmt.Printf("  %sAction     :%s %s\n", Gray, White, actionMode)
	fmt.Printf("  %sVersion    :%s %s\n", Gray, White, version)
	fmt.Printf("  %sOutput Dir :%s %s\n", Gray, White, binDir)
	fmt.Println()

	// PREPARE DIRS
	spinner, _ := pterm.DefaultSpinner.Start("Preparing workspace...")

	safePrepareDir(binDir)
	if copyNpm {
		safePrepareDir(NpmDir)
	}

	spinner.Success("Workspace cleaned & ready")

	ldflags := fmt.Sprintf("-s -w -X 'mockserver/pkg/appinfo.BuildDate=%s'", buildDate)

	// DETERMINE ACTIVE TARGETS
	var activeTargets []Target

	if buildAll {
		activeTargets = allTargets
	} else {
		// Current Mode
		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
		foundName := AppName + ext
		for _, t := range allTargets {
			if t.OS == runtime.GOOS && t.Arch == runtime.GOARCH {
				foundName = t.FileName
				break
			}
		}
		activeTargets = []Target{
			{runtime.GOOS, runtime.GOARCH, foundName},
		}
	}

	// BUILD LOOP
	printSection("\nCOMPILATION TASKS")

	for _, t := range activeTargets {
		outPath := filepath.Join(binDir, t.FileName)

		// Build Process
		dur, err := runBuildHybrid(t.OS, t.Arch, ldflags, outPath)

		statusStr := pterm.FgGreen.Sprint("SUCCESS")
		if err != nil {
			statusStr = pterm.FgRed.Sprint("FAILED")
		}

		results = append(results, BuildResult{
			Platform: fmt.Sprintf("%s/%s", t.OS, t.Arch),
			Status:   statusStr,
			Duration: dur,
			Artifact: t.FileName,
		})

		// NPM Copy
		if copyNpm && err == nil {
			npmPath := filepath.Join(NpmDir, t.FileName)
			if err := copyWithOverwrite(outPath, npmPath, "NPM Dist"); err != nil {
				fmt.Printf("    %s%s NPM Copy Failed: %v%s\n", Red, IconX, err, Reset)
			}
		}
	}

	if buildAll {
		printSection("\nLOCAL ENVIRONMENT")
		copyLocal(binDir)
	}

	printSummaryTable(time.Since(startTotal))
}

func runBuildHybrid(osName, arch, ldflags, outPath string) (time.Duration, error) {
	label := fmt.Sprintf("%s/%s", osName, arch)

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Building for %s...", label))
	start := time.Now()

	for {
		cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outPath, MainPkg)
		cmd.Env = append(os.Environ(), "GOOS="+osName, "GOARCH="+arch)

		output, err := cmd.CombinedOutput()

		if err == nil {
			dur := time.Since(start)
			spinner.Success(fmt.Sprintf("Building for %-15s", label))
			return dur, nil
		}

		spinner.Fail(fmt.Sprintf("Building for %s failed", label))
		fmt.Printf("    %s%s Error: %s%s\n", Red, IconX, strings.TrimSpace(string(output)), Reset)

		if askYesNo(fmt.Sprintf("    %s⚠️  Cannot write to '%s'. Overwrite (Force)?", Yellow, outPath)) {
			os.Remove(outPath)
			fmt.Printf("    %s%s Retrying build...%s\n", Cyan, IconArrow, Reset)
			spinner, _ = pterm.DefaultSpinner.Start(fmt.Sprintf("Retrying %s...", label))
			continue
		}

		fmt.Printf("    %s%s Skipped by user.%s\n", Yellow, IconX, Reset)
		return time.Since(start), err
	}
}

func copyLocal(outDir string) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}

	var srcName string
	for _, t := range allTargets {
		if t.OS == goos && t.Arch == goarch {
			srcName = t.FileName
			break
		}
	}

	if srcName == "" {
		return
	}

	srcPath := filepath.Join(outDir, srcName)
	destPath := filepath.Join(outDir, AppName+ext)

	if srcName == AppName+ext {
		return
	} // If it has the same name, it come out.

	if err := copyWithOverwrite(srcPath, destPath, "Local Binary"); err == nil {
		pterm.Success.Printf("Created local alias: %s -> %s\n", srcName, AppName+ext)
	} else {
		pterm.Error.Printf("Failed to create local alias: %s\n", destPath)
	}
}

func copyWithOverwrite(src, dst, label string) error {
	for {
		input, err := os.ReadFile(src)
		if err != nil {
			return err
		}

		err = os.WriteFile(dst, input, 0755)
		if err == nil {
			return nil
		}

		fmt.Printf("\n    %s%s [%s] Copy Error: %s%s\n", Red, IconX, label, err, Reset)

		if askYesNo(fmt.Sprintf("    %s⚠️  File '%s' is busy. Overwrite?", Yellow, dst)) {
			if rmErr := os.Remove(dst); rmErr != nil {
				fmt.Printf("      %s(Could not auto-delete, close app manually)%s\n", Red, Reset)
			} else {
				fmt.Printf("      %s(Old file deleted)%s\n", Green, Reset)
			}
			continue
		}
		return fmt.Errorf("skipped by user")
	}
}

func safePrepareDir(path string) {
	os.RemoveAll(path)
	os.MkdirAll(path, 0755)
}

// UI HELPERS
func askYesNo(question string) bool {
	fmt.Printf("%s [Y/n]: %s", question, Reset)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}

func printBanner() {
	fmt.Println()
	fmt.Printf("%s╔═══════════════════════════════════════════════════════════════╗%s\n", Blue, Reset)
	fmt.Printf("%s║                   MOCKSERVER BUILD ENGINE                     ║%s\n", Blue, Reset)
	fmt.Printf("%s╚═══════════════════════════════════════════════════════════════╝%s\n", Blue, Reset)
	fmt.Println()
}

func printSection(title string) {
	fmt.Printf("%s%s%s\n", Bold, title, Reset)
	fmt.Printf("%s%s%s\n", Gray, strings.Repeat("-", 65), Reset)
}

func printSummaryTable(totalDur time.Duration) {
	fmt.Println()
	printSection("BUILD SUMMARY")

	tableData := [][]string{
		{"PLATFORM", "STATUS", "DURATION", "ARTIFACT"},
	}

	for _, r := range results {
		durStr := fmt.Sprintf("%v", r.Duration.Round(time.Millisecond))
		tableData = append(tableData, []string{r.Platform, r.Status, durStr, r.Artifact})
	}

	pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()

	fmt.Println()
	fmt.Printf("%s✨ Total Time: %v%s\n", Cyan, totalDur.Round(time.Millisecond), Reset)
	fmt.Printf("%s📂 Artifacts : %s%s\n", Cyan, binDir, Reset)
	fmt.Println()
}
