package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	fsys "kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/filesystem"
)

const (
	// Git status
	FILE_ADDED   = "A"
	FILE_DELETED = "D"
	FILE_RENAMED = "R"

	TARGET_NONE = "none"
)

type parameters struct {
	tag       string
	rulesFile string
	debug     bool
}

type Exclude struct {
	Pattern string   `yaml:"pattern"`
	Exclude []string `yaml:"exclude"`
}

type Specific struct {
	Pattern string   `yaml:"pattern"`
	Targets []string `yaml:"targets"`
}

type Config struct {
	// All: all the VM-based providers will be provisioned
	All []string `yaml:"all"`
	// None: none of the VM-based providers will be provisioned
	None []string `yaml:"none"`
	// Regex: the regex will be globbed, and for each directory,
	// there will be a rule where the directory affects the specific provider:
	// a/b/k8s-X.YZ - X.YZ.
	// Notes:
	//   This rule is always recursive
	//   Only the last dir can be a regex
	Regex []string `yaml:"regex"`
	// RegexNone: The regex will be globbed,
	// and for each directory, there will be a rule where the directory affects none of the providers:
	// cluster-up/cluster/kind-X.YZ - none
	// Notes:
	//   key should be a directory ending with `*`
	RegexNone []string `yaml:"regex_none"`
	// Exclude: specific VM-based provider(s) that will be provisioned
	Exclude []Exclude `yaml:"exclude"`
	// Specific: all beside the given target(s) will be provisioned
	Specific []Specific `yaml:"specific"`
}

type OutputSplitter struct{}

// NewProvisionManagerCommand determines which providers should be rebuilt
func NewProvisionManagerCommand() *cobra.Command {
	provision := &cobra.Command{
		Use:   "provision-manager",
		Short: "provision manager determines which providers should be rebuilt",
		RunE:  provisionManager,
		Args:  cobra.ExactArgs(0),
	}
	provision.Flags().String("tag", "", "kubevirtci tag to compare to, default: fetch latest")
	provision.Flags().String("rules", "hack/pman/rules.yaml", "rules file")
	provision.Flags().Bool("debug", false, "run in debug mode, default: false")

	return provision
}

func provisionManager(cmd *cobra.Command, arguments []string) error {
	params, err := parseArguments(cmd)
	if err != nil {
		return err
	}

	configLogger(params.debug)

	// Sleep to let logrus flush its buffer, in order to avoid race between logrus and printing of the return value
	defer func() {
		time.Sleep(200 * time.Millisecond)
	}()

	if len(params.tag) == 0 {
		params.tag, err = getKubevirtciTag()
		if err != nil {
			return err
		}
	}

	printSection("Parameters")
	logrus.Debug("Tag: ", params.tag)

	targets, err := getTargets("cluster-provision/k8s/*")
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("No valid targets found")
	}

	rulesDB, err := buildRulesDBfromFile(params.rulesFile, targets)
	if err != nil {
		return err
	}

	targetToRebuild, err := processGitNameStatusChanges(rulesDB, targets, params.tag)
	if err != nil {
		return err
	}

	j, err := json.Marshal(targetToRebuild)
	if err != nil {
		return err
	}

	printSection("Result")
	fmt.Println(string(j))

	return nil
}

func parseArguments(cmd *cobra.Command) (parameters, error) {
	params := parameters{}
	var err error

	params.debug, err = cmd.Flags().GetBool("debug")
	if err != nil {
		return parameters{}, err
	}

	params.rulesFile, err = cmd.Flags().GetString("rules")
	if err != nil {
		return parameters{}, err
	}

	params.tag, err = cmd.Flags().GetString("tag")
	if err != nil {
		return parameters{}, err
	}

	return params, nil
}

func configLogger(debug bool) {
	logrus.SetOutput(&OutputSplitter{})

	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{DisableLevelTruncation: true, ForceColors: true, DisableTimestamp: true})
	}
}

func (splitter *OutputSplitter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("level=error")) || bytes.Contains(p, []byte("level=fatal")) {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}

func getTargets(path string) ([]string, error) {
	directories, err := fsys.GlobDirectories("cluster-provision/k8s/*")
	if err != nil {
		return nil, err
	}

	targets := []string{}
	for _, dir := range directories {
		targets = append(targets, filepath.Base(dir))
	}

	logrus.Debug("Targets: ", targets)
	return targets, nil
}

func processGitNameStatusChanges(rulesDB map[string][]string, targets []string, tag string) (map[string]bool, error) {
	cmdOutput, err := runCommand("git", []string{"diff", "--name-status", tag})
	if err != nil {
		return nil, err
	}

	return processChanges(rulesDB, targets, cmdOutput)
}

func processChanges(rulesDB map[string][]string, targets []string, changes string) (map[string]bool, error) {
	targetToRebuild := make(map[string]bool)
	for _, target := range targets {
		targetToRebuild[target] = false
	}

	printSection("Changed files")

	errorFound := false
	files := strings.Split(changes, "\n")
	for _, nameStatus := range files {
		if nameStatus == "" {
			break
		}

		tokens := strings.Split(nameStatus, "\t")
		if len(tokens) < 2 {
			return nil, fmt.Errorf("wrong input syntax, should be <status>\\t<filename>")
		}

		status := tokens[0]
		fileName := tokens[1]

		if strings.HasPrefix(status, FILE_RENAMED) {
			if len(tokens) != 3 {
				return nil, fmt.Errorf("wrong input syntax, should be <status>\\t<old_filename>\\t<new_filename>")
			}
			fileName = tokens[2]
		}

		// Skip markdown files
		if strings.HasSuffix(fileName, ".md") {
			continue
		}

		match, err := matcher(rulesDB, fileName, status)
		if err != nil {
			errorFound = true
			logrus.Error(err)
			continue
		}

		if !errorFound {
			logrus.Debug(status + " : " + fileName + " - [" + strings.Join(match, " ") + "]")
		}

		for _, target := range match {
			if target != TARGET_NONE {
				targetToRebuild[target] = true
			}
		}
	}

	if errorFound {
		return nil, fmt.Errorf("Errors detected: files dont have a matching rule")
	}

	return targetToRebuild, nil
}

func matcher(rulesDB map[string][]string, fileName string, status string) ([]string, error) {
	match, ok := rulesDB[fileName]
	if ok {
		return match, nil
	}

	match, ok = rulesDB[filepath.Dir(fileName)]
	if ok {
		return match, nil
	}

	candid := fileName
	for candid != "." && candid != "/" {
		candid = filepath.Dir(candid)
		match, ok = rulesDB[candid+"/*"]
		if ok {
			return match, nil
		}
	}

	if status != FILE_DELETED {
		return nil, fmt.Errorf("Failed to find a rule for " + fileName)
	}

	return nil, nil
}

func getKubevirtciTag() (string, error) {
	const kubevirtciTagUrl = "https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest?ignoreCache=1"

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(kubevirtciTagUrl)
	if err != nil {
		return "", fmt.Errorf("ERROR: getting latest kubevirtci tag failed, error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ERROR: getting latest kubevirtci tag failed, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ERROR: parsing response body failed: %v", err)
	}

	return strings.TrimSuffix(string(body), "\n"), nil
}

func runCommand(command string, args []string) (string, error) {
	var stderr1 bytes.Buffer
	remoteAddCmd := exec.Command("git", "remote", "add", "upstream", "https://github.com/kubevirt/kubevirtci.git")
	remoteAddCmd.Stderr = &stderr1
	if err := remoteAddCmd.Run(); err != nil {
		logrus.Debugf("Failed to add remote named upstream: %v, stderr: %s", err, stderr1.String()) //upstream remote already may exist
	} else {
		logrus.Debug("Successfully added remote named upstream.")
	}

	var stderr2 bytes.Buffer
	fetchCmd := exec.Command("git", "fetch", "upstream", "--tags")
	fetchCmd.Stderr = &stderr2
	if err := fetchCmd.Run(); err != nil {
		logrus.Debugf("Failed to fetch tags from upstream: %v, stderr: %s", err, stderr2.String())
	}

	var stdout3, stderr3 bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout3
	cmd.Stderr = &stderr3
	err := cmd.Run()
	if err != nil {
		if strings.Contains(stderr3.String(), "unknown revision or path not in the working tree") {
			logrus.Error("Tag not found, please run 'git fetch upstream --tags'")
		}
		return "", errors.Wrapf(err, "Failed to run command: %s %s\nStdout:\n%s\nStderr:\n%s",
			command, strings.Join(args, " "), cmd.Stdout, cmd.Stderr)
	}

	return stdout3.String(), nil
}

func buildRulesDBfromFile(rulesFile string, targets []string) (map[string][]string, error) {
	inFile, err := fsys.GetFileSystem().Open(rulesFile)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	return buildRulesDB(inFile, targets)
}

func buildRulesDB(input io.Reader, targets []string) (map[string][]string, error) {
	rulesDB := make(map[string][]string)

	buffer, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(buffer, &cfg)
	if err != nil {
		return nil, err
	}

	err = addAllRules(cfg.All, targets, rulesDB)
	if err != nil {
		return nil, err
	}
	err = addNoneRules(cfg.None, rulesDB)
	if err != nil {
		return nil, err
	}
	err = addRegexRules(cfg.Regex, targets, rulesDB)
	if err != nil {
		return nil, err
	}
	err = addRegexNoneRules(cfg.RegexNone, rulesDB)
	if err != nil {
		return nil, err
	}
	err = addExcludeRules(cfg.Exclude, targets, rulesDB)
	if err != nil {
		return nil, err
	}
	err = addSpecificRules(cfg.Specific, targets, rulesDB)
	if err != nil {
		return nil, err
	}

	printRulesDB(rulesDB)
	return rulesDB, nil
}

func addAllRules(cfgAll []string, targets []string, rulesDB map[string][]string) error {
	for _, path := range cfgAll {
		err := validatePathExist(path, "All")
		if err != nil {
			return err
		}
		rulesDB[path] = targets
	}

	return nil
}

func addNoneRules(cfgNone []string, rulesDB map[string][]string) error {
	for _, path := range cfgNone {
		err := validatePathExist(path, "None")
		if err != nil {
			return err
		}
		rulesDB[path] = []string{TARGET_NONE}
	}

	return nil
}

func addRegexRules(cfgRegex []string, targets []string, rulesDB map[string][]string) error {
	for _, path := range cfgRegex {
		directories, _ := fsys.GlobDirectories(path)
		if len(directories) == 0 {
			return fmt.Errorf("No valid directories found for Regex rule: " + path)
		}
		for _, dir := range directories {
			target := strings.ReplaceAll(filepath.Base(dir), "k8s-", "")
			if !isTargetValid(target, targets) {
				return fmt.Errorf("Invalid target " + target + ", regex rule: " + path)
			}
			rulesDB[dir+"/*"] = []string{target}
		}
	}

	return nil
}

func addRegexNoneRules(cfgRegexNone []string, rulesDB map[string][]string) error {
	for _, path := range cfgRegexNone {
		directories, _ := fsys.GlobDirectories(path)
		if len(directories) == 0 {
			return fmt.Errorf("No valid directories found for RegexNone rule: " + path)
		}
		for _, dir := range directories {
			rulesDB[dir+"/*"] = []string{TARGET_NONE}
		}
	}

	return nil
}

func addExcludeRules(cfgExclude []Exclude, targets []string, rulesDB map[string][]string) error {
	for _, e := range cfgExclude {
		err := validatePathExist(e.Pattern, "Exclude")
		if err != nil {
			return err
		}
		ruleTargets := append([]string(nil), targets...)
		for _, target := range e.Exclude {
			if !isTargetValid(target, targets) {
				return fmt.Errorf("Invalid target, exclude rule: " + target)
			}
			ruleTargets = excludeTarget(target, ruleTargets)
		}
		if len(ruleTargets) == 0 {
			return fmt.Errorf("No valid targets left for exclude rule: " + e.Pattern)
		}
		rulesDB[e.Pattern] = ruleTargets
	}
	return nil
}

func addSpecificRules(cfgSpecific []Specific, targets []string, rulesDB map[string][]string) error {
	for _, e := range cfgSpecific {
		err := validatePathExist(e.Pattern, "Specific")
		if err != nil {
			return err
		}
		ruleTargets := []string{}
		for _, target := range e.Targets {
			if !isTargetValid(target, targets) {
				return fmt.Errorf("Invalid target, specific rule: " + target)
			}
			ruleTargets = append(ruleTargets, target)
		}
		rulesDB[e.Pattern] = ruleTargets
	}
	return nil
}

func validatePathExist(path string, ruleType string) error {
	matches, err := fsys.GetFileSystem().Glob(path)
	if err != nil {
		return fmt.Errorf("Error occurred for rule %s %s: %v", ruleType, path, err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("Path doesn't exist for rule %s: %s", ruleType, path)
	}
	return nil
}

func printRulesDB(rulesDB map[string][]string) {
	keys := make([]string, 0, len(rulesDB))
	for k := range rulesDB {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	printSection("Rules")

	for _, k := range keys {
		logrus.Debug(k + " : [" + strings.Join(rulesDB[k], " ") + "]")
	}
}

func isTargetValid(target string, targets []string) bool {
	for _, t := range targets {
		if t == target {
			return true
		}
	}
	return false
}

func excludeTarget(target string, targets []string) []string {
	newTargets := []string{}
	for _, t := range targets {
		if t != target {
			newTargets = append(newTargets, t)
		}
	}
	return newTargets
}

func printSection(title string) {
	title += ":"
	dashes := strings.Repeat("-", len(title))

	logrus.Debug(dashes)
	logrus.Debug(title)
	logrus.Debug(dashes)
}
