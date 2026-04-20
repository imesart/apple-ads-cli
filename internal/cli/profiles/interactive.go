package profiles

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
	"golang.org/x/term"
)

var (
	stdinIsTerminal     = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
	stdoutIsTerminal    = func() bool { return term.IsTerminal(int(os.Stdout.Fd())) }
	stderrIsTerminal    = func() bool { return term.IsTerminal(int(os.Stderr.Fd())) }
	openURLFunc         = openURL
	openPrivateURLFunc  = openPrivateURL
	copyToClipboardFunc = copyToClipboard
)

const instructionColor = "\x1b[38;5;110m"
const ansiReset = "\x1b[0m"

type browserLaunch struct {
	Name    string
	Command []string
	Check   func() bool
}

// SetInteractiveFuncsForTesting overrides terminal/browser helpers for tests.
func SetInteractiveFuncsForTesting(
	stdinTTY func() bool,
	stdoutTTY func() bool,
	stderrTTY func() bool,
	openURL func(string) error,
	openPrivateURL func(string) error,
	copyToClipboard func(string) error,
) func() {
	prevStdinTTY := stdinIsTerminal
	prevStdoutTTY := stdoutIsTerminal
	prevStderrTTY := stderrIsTerminal
	prevOpenURL := openURLFunc
	prevOpenPrivateURL := openPrivateURLFunc
	prevCopyToClipboard := copyToClipboardFunc
	if stdinTTY != nil {
		stdinIsTerminal = stdinTTY
	}
	if stdoutTTY != nil {
		stdoutIsTerminal = stdoutTTY
	}
	if stderrTTY != nil {
		stderrIsTerminal = stderrTTY
	}
	if openURL != nil {
		openURLFunc = openURL
	}
	if openPrivateURL != nil {
		openPrivateURLFunc = openPrivateURL
	}
	if copyToClipboard != nil {
		copyToClipboardFunc = copyToClipboard
	}
	return func() {
		stdinIsTerminal = prevStdinTTY
		stdoutIsTerminal = prevStdoutTTY
		stderrIsTerminal = prevStderrTTY
		openURLFunc = prevOpenURL
		openPrivateURLFunc = prevOpenPrivateURL
		copyToClipboardFunc = prevCopyToClipboard
	}
}

const inviteAPIUserInstructions = `Invite an API user in Apple Ads:
- Sign in as an account administrator, or ask an administrator to do these steps.
- Open User Management for the right organization.
- Press Invite Users.
- Assign API Account Manager or API Account Read Only.`

const setupAPIAccessInstructions = `Set up API access in Apple Ads:
- Open the email invitation link if a new API user was invited.
- Sign in with an API-enabled Apple Ads account.
- Open API Account Settings for the right organization.
- Paste the public key, including the BEGIN/END PUBLIC KEY lines.
- Press Save.
- Copy the clientId, teamId, and keyId values.`

type promptSession struct {
	in  *bufio.Reader
	out io.Writer
}

func runInteractiveCreate(ctx context.Context, inputs createCommandInputs, check bool, output shared.OutputFlags) error {
	if !stdinIsTerminal() || !stdoutIsTerminal() {
		return shared.UsageError("--interactive requires a terminal on stdin and stdout")
	}

	session := &promptSession{
		in:  bufio.NewReader(os.Stdin),
		out: os.Stderr,
	}
	var err error
	if inputs.Name == "" {
		inputs.Name, err = session.promptString("Profile name", "default")
		if err != nil {
			return err
		}
	}
	if inputs.Name == "" {
		inputs.Name = "default"
	}

	cf := config.LoadFile()
	if _, exists := cf.Profiles[inputs.Name]; exists {
		return shared.ReportError(fmt.Errorf("profile %q already exists; use 'aads profiles update' to modify it", inputs.Name))
	}

	if inputs.PrivateKeyPath == "" {
		inputs.PrivateKeyPath = defaultPrivateKeyPath(inputs.Name)
	}
	if inputs.ExplicitPrivateKey {
		if _, err := os.Stat(expandUserPath(inputs.PrivateKeyPath)); os.IsNotExist(err) {
			return shared.ReportError(fmt.Errorf("private key file %q does not exist", inputs.PrivateKeyPath))
		} else if err != nil {
			return shared.ReportError(fmt.Errorf("checking private key path %q: %w", inputs.PrivateKeyPath, err))
		}
	} else if _, err := os.Stat(expandUserPath(inputs.PrivateKeyPath)); os.IsNotExist(err) {
		if err := session.ensureDefaultPrivateKey(ctx, &inputs, check); err != nil {
			return err
		}
	} else if err != nil {
		return shared.ReportError(fmt.Errorf("checking private key path %q: %w", inputs.PrivateKeyPath, err))
	}

	if inputs.ClientID == "" || inputs.TeamID == "" || inputs.KeyID == "" {
		if err := session.collectCredentials(&inputs); err != nil {
			return err
		}
	}

	discovered := discoveredProfileDefaults{}
	if !inputs.ExplicitOrgID || !inputs.ExplicitDefaultCurrency || inputs.DefaultTimezone == "" {
		discovered, err = discoverProfileDefaults(ctx, profileCredentialSubset(inputs), "")
		if err != nil && !inputs.ExplicitOrgID {
			fmt.Fprintf(session.out, "Warning: could not inspect Apple Ads orgs data (ACLs) before creation: %v\n", err)
		}
	}
	for _, warning := range discovered.Warnings {
		fmt.Fprintf(session.out, "Warning: %s\n", warning)
	}

	if !inputs.ExplicitOrgID {
		if err := session.resolveInteractiveOrg(&inputs, discovered); err != nil {
			return err
		}
	}
	applyDiscoveredDefaults(&inputs, discovered)
	if !inputs.ExplicitOrgID {
		applySelectedOrgDefaults(&inputs, discovered)
	}

	if !inputs.ExplicitOrgID && strings.TrimSpace(inputs.OrgID) == "" {
		for strings.TrimSpace(inputs.OrgID) == "" {
			orgID, err := session.promptString("Organization ID", "")
			if err != nil {
				return err
			}
			inputs.OrgID = strings.TrimSpace(orgID)
			if inputs.OrgID == "" {
				fmt.Fprintln(session.out, "Organization ID is required.")
			}
		}
	}

	if !inputs.ExplicitDefaultCurrency {
		currency, err := session.promptStringWithNote("Default currency", strings.TrimSpace(inputs.DefaultCurrency), "Press Enter to keep the detected value or leave it blank.")
		if err != nil {
			return err
		}
		inputs.DefaultCurrency = strings.TrimSpace(currency)
	}

	if inputs.MaxDailyBudget == "" {
		value, err := session.promptStringWithNote("Max daily budget", "0", "0 means no limit.")
		if err != nil {
			return err
		}
		inputs.MaxDailyBudget = strings.TrimSpace(value)
	}
	if inputs.MaxBid == "" {
		value, err := session.promptStringWithNote("Max bid", "0", "0 means no limit.")
		if err != nil {
			return err
		}
		inputs.MaxBid = strings.TrimSpace(value)
	}
	if inputs.MaxCPAGoal == "" {
		value, err := session.promptStringWithNote("Max CPA goal", "0", "0 means no limit.")
		if err != nil {
			return err
		}
		inputs.MaxCPAGoal = strings.TrimSpace(value)
	}

	return runCreateFlow(ctx, inputs, check, output)
}

func profileCredentialSubset(inputs createCommandInputs) config.Profile {
	return config.Profile{
		ClientID:       strings.TrimSpace(inputs.ClientID),
		TeamID:         strings.TrimSpace(inputs.TeamID),
		KeyID:          strings.TrimSpace(inputs.KeyID),
		PrivateKeyPath: strings.TrimSpace(inputs.PrivateKeyPath),
	}
}

func applyDiscoveredDefaults(inputs *createCommandInputs, discovered discoveredProfileDefaults) {
	if inputs.DefaultCurrency == "" {
		inputs.DefaultCurrency = discovered.DefaultCurrency
	}
	if inputs.DefaultTimezone == "" {
		inputs.DefaultTimezone = discovered.DefaultTimezone
	}
}

func applySelectedOrgDefaults(inputs *createCommandInputs, discovered discoveredProfileDefaults) {
	for _, org := range discovered.Organizations {
		if org.OrgID != strings.TrimSpace(inputs.OrgID) {
			continue
		}
		if inputs.DefaultCurrency == "" {
			inputs.DefaultCurrency = org.Currency
		}
		if inputs.DefaultTimezone == "" {
			inputs.DefaultTimezone = org.Timezone
		}
		return
	}
}

func (s *promptSession) promptString(label, defaultValue string) (string, error) {
	return s.promptStringWithNote(label, defaultValue, "")
}

func (s *promptSession) promptStringWithNote(label, defaultValue, note string) (string, error) {
	if note != "" {
		fmt.Fprintln(s.out, note)
	}
	if defaultValue != "" {
		fmt.Fprintf(s.out, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(s.out, "%s: ", label)
	}
	value, err := s.readLine()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	return strings.TrimSpace(value), nil
}

func (s *promptSession) promptYesNo(label string, defaultYes bool) (bool, error) {
	choice := "y/N"
	if defaultYes {
		choice = "Y/n"
	}
	fmt.Fprintf(s.out, "%s [%s]: ", label, choice)
	value, err := s.readLine()
	if err != nil {
		return false, err
	}
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return defaultYes, nil
	}
	switch value {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		fmt.Fprintln(s.out, "Enter y or n.")
		return s.promptYesNo(label, defaultYes)
	}
}

func (s *promptSession) readLine() (string, error) {
	line, err := s.in.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (s *promptSession) waitForEnter(prompt string) error {
	fmt.Fprintf(s.out, "%s", prompt)
	_, err := s.readLine()
	return err
}

func (s *promptSession) ensureDefaultPrivateKey(ctx context.Context, inputs *createCommandInputs, check bool) error {
	hasAPIAccess, err := s.promptYesNo("Do you already have an Apple Ads account with API Account Manager or API Account Read Only access?", false)
	if err != nil {
		return err
	}
	hasPrivateKey := false
	if !hasAPIAccess {
		fmt.Fprintln(s.out)
		fmt.Fprintln(s.out, styleInstruction(inviteAPIUserInstructions))
		fmt.Fprintln(s.out)
		if check {
			fmt.Fprintln(s.out, "Skipping browser launch because --check is set. Open https://app-ads.apple.com/cm/app/settings/users manually if needed.")
		} else {
			if err := s.waitForEnter("Press Enter to open Apple Ads User Management in your browser."); err != nil {
				return err
			}
			if err := openURLFunc("https://app-ads.apple.com/cm/app/settings/users"); err != nil {
				fmt.Fprintf(s.out, "Warning: could not open browser automatically: %v\n", err)
				fmt.Fprintln(s.out, "Open https://app-ads.apple.com/cm/app/settings/users manually.")
			}
			fmt.Fprintln(s.out)
			if err := s.waitForEnter("Press Enter to continue after the invitation is sent."); err != nil {
				return err
			}
		}
	} else {
		hasPrivateKey, err = s.promptYesNo("Do you already have the private key PEM file for this profile?", false)
		if err != nil {
			return err
		}
	}
	if hasPrivateKey {
		path, err := s.promptString("Private key path", inputs.PrivateKeyPath)
		if err != nil {
			return err
		}
		path = strings.TrimSpace(path)
		if path == "" {
			return shared.ValidationError("private key path is required when reusing an existing PEM file")
		}
		if _, err := os.Stat(expandUserPath(path)); os.IsNotExist(err) {
			return shared.ReportError(fmt.Errorf("private key file %q does not exist", path))
		} else if err != nil {
			return shared.ReportError(fmt.Errorf("checking private key path %q: %w", path, err))
		}
		inputs.PrivateKeyPath = path
		inputs.ExplicitPrivateKey = true
		return nil
	}

	if check {
		fmt.Fprintf(s.out, "Skipping key generation because --check is set. The default key path is %q.\n", inputs.PrivateKeyPath)
		return nil
	}
	if err := generatePrivateKey(ctx, inputs.PrivateKeyPath); err != nil {
		return shared.ReportError(err)
	}
	publicKey, err := publicKeyFromPrivateKey(inputs.PrivateKeyPath)
	if err != nil {
		return shared.ReportError(err)
	}
	fmt.Fprintln(s.out)
	fmt.Fprintln(s.out, styleInstruction(setupAPIAccessInstructions))
	fmt.Fprintln(s.out)
	fmt.Fprintln(s.out, styleInstruction("Public key to paste into Apple Ads:"))
	fmt.Fprintln(s.out, styleInstruction(publicKey))
	if err := copyToClipboardFunc(publicKey); err == nil {
		fmt.Fprintln(s.out, styleInstruction("Public key copied to clipboard."))
	}
	privateBrowser := firstAvailablePrivateBrowser()
	prompt := "Press Enter to open Apple Ads API settings in a private browser window if available."
	if privateBrowser.Name != "" {
		prompt = fmt.Sprintf("Press Enter to open Apple Ads API settings in %s private browsing.", privateBrowser.Name)
	}
	if err := s.waitForEnter(prompt); err != nil {
		return err
	}
	if err := openPrivateURLFunc("https://app-ads.apple.com/cm/app/settings/apicertificates"); err != nil {
		fmt.Fprintf(s.out, "Warning: could not open a private browser window automatically: %v\n", err)
		fmt.Fprintln(s.out, "Open https://app-ads.apple.com/cm/app/settings/apicertificates manually.")
	}
	return nil
}

func (s *promptSession) collectCredentials(inputs *createCommandInputs) error {
	fmt.Fprintln(s.out)
	fmt.Fprintln(s.out, styleInstruction("Paste the Apple Ads credential block with clientId, teamId, and keyId. Submit an empty line when finished."))
	block, err := s.readMultilineBlock()
	if err != nil {
		return err
	}
	clientID, teamID, keyID := parseCredentialBlock(block)
	if inputs.ClientID == "" {
		inputs.ClientID = clientID
	}
	if inputs.TeamID == "" {
		inputs.TeamID = teamID
	}
	if inputs.KeyID == "" {
		inputs.KeyID = keyID
	}
	if inputs.ClientID == "" {
		value, err := s.promptSearchADSValue("clientId")
		if err != nil {
			return err
		}
		inputs.ClientID = value
	}
	if inputs.TeamID == "" {
		value, err := s.promptSearchADSValue("teamId")
		if err != nil {
			return err
		}
		inputs.TeamID = value
	}
	if inputs.KeyID == "" {
		value, err := s.promptString("keyId", "")
		if err != nil {
			return err
		}
		inputs.KeyID = value
	}
	return nil
}

func styleInstruction(text string) string {
	if shared.NoColor() || !stderrIsTerminal() {
		return text
	}
	return instructionColor + text + ansiReset
}

func copyToClipboard(text string) error {
	cmd := clipboardCommand()
	if len(cmd) == 0 {
		return fmt.Errorf("clipboard command not available")
	}
	if _, err := exec.LookPath(cmd[0]); err != nil {
		return err
	}
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = strings.NewReader(text)
	return c.Run()
}

func clipboardCommand() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"pbcopy"}
	case "windows":
		return []string{"clip"}
	default:
		if _, err := exec.LookPath("wl-copy"); err == nil {
			return []string{"wl-copy"}
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return []string{"xclip", "-selection", "clipboard"}
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			return []string{"xsel", "--clipboard", "--input"}
		}
		return nil
	}
}

func (s *promptSession) readMultilineBlock() (string, error) {
	var lines []string
	for {
		line, err := s.readLine()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), nil
}

func (s *promptSession) promptSearchADSValue(label string) (string, error) {
	for {
		value, err := s.promptString(label, "")
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(value, "SEARCHADS.") {
			return value, nil
		}
		fmt.Fprintf(s.out, "Warning: %s usually starts with SEARCHADS.\n", label)
		ok, err := s.promptYesNo(fmt.Sprintf("Use %q anyway?", value), false)
		if err != nil {
			return "", err
		}
		if ok {
			return value, nil
		}
	}
}

func (s *promptSession) resolveInteractiveOrg(inputs *createCommandInputs, discovered discoveredProfileDefaults) error {
	if len(discovered.Organizations) <= 1 {
		if inputs.OrgID == "" {
			inputs.OrgID = discovered.OrgID
		}
		return nil
	}
	fmt.Fprintln(s.out, "Organizations available from Apple Ads orgs list:")
	defaultChoice := ""
	for idx, org := range discovered.Organizations {
		fmt.Fprintf(s.out, "%d. %s (%s", idx+1, org.OrgID, org.OrgName)
		if org.Currency != "" {
			fmt.Fprintf(s.out, ", %s", org.Currency)
		}
		if org.Timezone != "" {
			fmt.Fprintf(s.out, ", %s", org.Timezone)
		}
		fmt.Fprintln(s.out, ")")
		if discovered.ParentOrgID != "" && discovered.ParentOrgID == org.OrgID {
			defaultChoice = org.OrgID
		}
	}
	for {
		prompt := "Organization ID or list number"
		if defaultChoice != "" {
			prompt += " [" + defaultChoice + "]"
		}
		fmt.Fprintf(s.out, "%s: ", prompt)
		value, err := s.readLine()
		if err != nil {
			return err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			inputs.OrgID = defaultChoice
			return nil
		}
		if index, err := strconv.Atoi(value); err == nil {
			if index >= 1 && index <= len(discovered.Organizations) {
				inputs.OrgID = discovered.Organizations[index-1].OrgID
				return nil
			}
		}
		for _, org := range discovered.Organizations {
			if value == org.OrgID {
				inputs.OrgID = value
				return nil
			}
		}
		fmt.Fprintln(s.out, "Enter one of the listed organization IDs or list numbers.")
	}
}

func parseCredentialBlock(block string) (clientID string, teamID string, keyID string) {
	lines := strings.Split(block, "\n")
	for idx := 0; idx < len(lines); idx++ {
		label := strings.TrimSpace(lines[idx])
		if idx+1 >= len(lines) {
			continue
		}
		value := strings.TrimSpace(lines[idx+1])
		switch strings.ToLower(label) {
		case "clientid":
			clientID = value
		case "teamid":
			teamID = value
		case "keyid":
			keyID = value
		}
	}
	if clientID != "" && teamID != "" && keyID != "" {
		return clientID, teamID, keyID
	}
	clientMatch := regexp.MustCompile(`(?i)clientId\s*[:\n\r ]+\s*([^\s]+)`).FindStringSubmatch(block)
	teamMatch := regexp.MustCompile(`(?i)teamId\s*[:\n\r ]+\s*([^\s]+)`).FindStringSubmatch(block)
	keyMatch := regexp.MustCompile(`(?i)keyId\s*[:\n\r ]+\s*([^\s]+)`).FindStringSubmatch(block)
	if clientID == "" && len(clientMatch) == 2 {
		clientID = strings.TrimSpace(clientMatch[1])
	}
	if teamID == "" && len(teamMatch) == 2 {
		teamID = strings.TrimSpace(teamMatch[1])
	}
	if keyID == "" && len(keyMatch) == 2 {
		keyID = strings.TrimSpace(keyMatch[1])
	}
	return clientID, teamID, keyID
}

func openURL(url string) error {
	return openWithCommand(defaultBrowserCommand(url))
}

func openPrivateURL(url string) error {
	for _, browser := range privateBrowserLaunches(url) {
		if browser.Check != nil && !browser.Check() {
			continue
		}
		if err := openWithCommand(browser.Command); err == nil {
			return nil
		}
	}
	return openURL(url)
}

func openWithCommand(command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("no browser command available")
	}
	if _, err := exec.LookPath(command[0]); err != nil {
		return err
	}
	cmd := exec.Command(command[0], command[1:]...)
	return cmd.Start()
}

func defaultBrowserCommand(url string) []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"open", url}
	case "windows":
		return []string{"rundll32", "url.dll,FileProtocolHandler", url}
	default:
		return []string{"xdg-open", url}
	}
}

func firstAvailablePrivateBrowser() browserLaunch {
	for _, browser := range privateBrowserLaunches("https://example.com") {
		if browser.Check != nil {
			if browser.Check() {
				return browser
			}
			continue
		}
		if len(browser.Command) == 0 {
			continue
		}
		if _, err := exec.LookPath(browser.Command[0]); err == nil {
			return browser
		}
	}
	return browserLaunch{}
}

func privateBrowserLaunches(url string) []browserLaunch {
	switch runtime.GOOS {
	case "darwin":
		var commands []browserLaunch
		for _, browser := range []struct {
			appName string
			name    string
			check   string
			flag    string
		}{
			{appName: "Google Chrome", name: "Google Chrome", check: "/Applications/Google Chrome.app", flag: "--incognito"},
			{appName: "Firefox", name: "Firefox", check: "/Applications/Firefox.app", flag: "--private-window"},
			{appName: "Safari", name: "Safari", check: "/Applications/Safari.app", flag: ""},
		} {
			command := []string{"open", "-na", browser.appName}
			if browser.flag != "" {
				command = append(command, "--args", browser.flag, url)
			} else {
				command = []string{
					"osascript",
					"-e", `tell application "Safari" to activate`,
					"-e", `tell application "System Events" to keystroke "n" using {command down, shift down}`,
					"-e", `delay 0.2`,
					"-e", fmt.Sprintf(`tell application "Safari" to open location %q`, url),
				}
			}
			checkPath := browser.check
			commands = append(commands, browserLaunch{
				Name:    browser.name,
				Command: command,
				Check: func() bool {
					_, err := os.Stat(checkPath)
					return err == nil
				},
			})
		}
		return commands
	case "windows":
		return []browserLaunch{
			{Name: "Chrome", Command: []string{"chrome", "--incognito", url}},
			{Name: "Firefox", Command: []string{"firefox", "--private-window", url}},
			{Name: "Microsoft Edge", Command: []string{"msedge", "--inprivate", url}},
		}
	default:
		return []browserLaunch{
			{Name: "google-chrome", Command: []string{"google-chrome", "--incognito", url}},
			{Name: "firefox", Command: []string{"firefox", "--private-window", url}},
			{Name: "microsoft-edge", Command: []string{"microsoft-edge", "--inprivate", url}},
		}
	}
}
