// Script aiming at checking the progress of the reference implementation,
// using Git.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

var (
	referenceGitDirectory = ""
	originCommitID        = "3d48bfc18731e3c2187a5b0666a7e94dcab0150b" // last commit "merged"
)

const remoteReferenceURL = "https://github.com/harfbuzz/harfbuzz"

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// parse the command line
func setupOptions() {
	refGit := flag.String("ref", "", "Git directory of the HarfBuzz reference implementation")
	origin := flag.String("origin", originCommitID, "CurrentCommit ID of the port")
	flag.Parse()

	if *refGit == "" {
		flag.Usage()
		os.Exit(1)
	}

	referenceGitDirectory = *refGit
	originCommitID = *origin
}

func getCommitsSince(originCommitID string) ([]string, error) {
	cmd := exec.Command("git", "rev-list", "--reverse", fmt.Sprintf("%s..HEAD", originCommitID))
	cmd.Dir = referenceGitDirectory
	commits, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(commits)), "\n"), nil
}

func getFilesFromCommit(commitID string) ([]string, error) {
	cmd := exec.Command("git", "diff-tree", "-r", "--name-only", "--no-commit-id", commitID)
	cmd.Dir = referenceGitDirectory
	files, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(files)), "\n"), nil
}

var regexpKind = regexp.MustCompile(`\[(\w+)\]`)

// return the topic of the commit or an empty string if not found
func getCommitKind(commitID string) (string, error) {
	cmd := exec.Command("git", "log", "-n 1", "--format=%s", commitID)
	cmd.Dir = referenceGitDirectory
	subject, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if m := regexpKind.FindSubmatch(subject); m != nil {
		return string(m[1]), nil
	}
	return "", nil
}

type index []string

func (i index) isTrackingFile(file string) bool {
	for _, trackFileOrDir := range i {
		if strings.HasPrefix(file, trackFileOrDir) {
			return true
		}
	}
	return false
}

// parse the reference index
func getTrackedFiles() (index, error) {
	b, err := ioutil.ReadFile("references_files.txt")
	if err != nil {
		return nil, err
	}
	filesOrDirs := strings.Split(string(b), "\n")
	return filesOrDirs, nil
}

type commitFiles struct {
	commitID     string
	trackedFiles []string // included in the reference index
	ignoredFiles []string // other files
}

func listChangedFiles() ([]commitFiles, int, []string) {
	trackedFiles, err := getTrackedFiles()
	check(err)

	commits, err := getCommitsSince(originCommitID)
	check(err)

	var (
		out          []commitFiles
		nbIgnored    int
		ignoredFiles = map[string]bool{}
	)
	for _, commit := range commits {
		files, err := getFilesFromCommit(commit)
		check(err)

		cf := commitFiles{commitID: commit}
		for _, file := range files {
			if trackedFiles.isTrackingFile(file) {
				cf.trackedFiles = append(cf.trackedFiles, file)
			} else {
				cf.ignoredFiles = append(cf.ignoredFiles, file)
			}
		}

		kind, err := getCommitKind(commit)
		check(err)

		// we simply ignore the commits with no tracked files
		// we also ignore the commits concerning the subset feature
		if len(cf.trackedFiles) == 0 || kind == "subset" {
			for _, file := range cf.ignoredFiles {
				ignoredFiles[file] = true
			}
			nbIgnored++
			continue
		}
		out = append(out, cf)
	}

	var ignored []string
	for f := range ignoredFiles {
		ignored = append(ignored, f)
	}
	sort.Strings(ignored)
	return out, nbIgnored, ignored
}

func main() {
	setupOptions()

	changes, nbIgnored, ignoredFiles := listChangedFiles()

	fmt.Printf("%d commits ignored since origin :\n\n", nbIgnored)
	fmt.Println(strings.Join(ignoredFiles, "\n"))
	fmt.Println()

	fmt.Printf("%d commits since origin :\n\n", len(changes))
	for _, change := range changes {
		url := remoteReferenceURL + "/commit/" + change.commitID
		fmt.Printf("commit %s : (%d) files ignored\n", url, len(change.ignoredFiles))
		for _, file := range change.trackedFiles {
			fmt.Println("\t", file)
		}
	}
}
