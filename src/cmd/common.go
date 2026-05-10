package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// passwordPromptSentinel is set as NoOptDefVal on --password flags so that bare
// --password (no value) signals "prompt interactively" rather than expecting
// the next argument as the password value.
// Side effect of NoOptDefVal: passing a value with a space (--password mypass) does NOT work —
// the value is silently treated as a positional argument. Users must use --password=mypass.
const passwordPromptSentinel = "\x00"

func completeKeyArg(toComplete string, matchType services.MatchType) ([]cobra.Completion, cobra.ShellCompDirective) {
	config := common.GetConfig()

	// If key has @ auto-complete DB name instead of key
	if key, db, found := strings.Cut(toComplete, "@"); found {
		var matchingDBs []string
		for registeredDB := range config.DBs {
			if strings.Contains(registeredDB, db) {
				matchingDBs = append(matchingDBs, key+"@"+registeredDB)
			}
		}

		return []cobra.Completion(matchingDBs), cobra.ShellCompDirectiveNoFileComp
	}

	// Otherwise, match keys
	var matchingKeys []string

	for db := range config.DBs {
		services.RunInTransaction(db, func(tx *sql.Tx) {
			matches := services.SearchKeys(tx, toComplete, matchType)
			if db != config.CurrentDB {
				for i := range matches {
					matches[i] += "@" + db
				}
			}

			matchingKeys = append(matchingKeys, matches...)
		})
	}

	return []cobra.Completion(matchingKeys), cobra.ShellCompDirectiveNoFileComp
}

// readPassword returns the password for cmd's --password flag.
// If provided with a value, returns it directly.
// If provided without a value (bare --password), prompts interactively with hidden input.
// When confirm is true (write operations), the prompt is shown twice and values must match.
func readPassword(cmd *cobra.Command, confirm bool) string {
	flag := cmd.Flags().Lookup("password")
	if flag == nil {
		panic("--password flag not defined")
	}

	// Sentinel means bare --password was passed; fall through to prompt
	val := flag.Value.String()
	if val != "" && val != passwordPromptSentinel {
		return val
	}

	return promptPassword(confirm)
}

func promptPassword(confirm bool) string {
	password := readPasswordFromTerminal("Password")

	if !confirm {
		return password
	}

	confirmed := readPasswordFromTerminal("Confirm password")

	if password != confirmed {
		common.Fail("Passwords do not match")
	}

	return password
}

func readPasswordFromTerminal(label string) string {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))

	// Clear the prompt line after entry
	fmt.Fprint(os.Stderr, "\r\033[K")

	common.FailOn(err)
	return string(password)
}
