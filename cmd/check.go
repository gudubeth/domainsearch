package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ozgio/dmn/pkg/dev"
	"github.com/ozgio/dmn/pkg/input"
	"github.com/ozgio/dmn/pkg/whois"
	"github.com/spf13/cobra"
)

type checkCmdFlags struct {
	showWhois bool
	benchmark bool
	parallel  int
}

func NewCheckCommand() *cobra.Command {
	var flags checkCmdFlags

	var cmd = &cobra.Command{
		Use:   "check",
		Short: "Check the vailability of domains",
		Long: `|
Checks the availability of domains. It needs full domain names. If you want
extended search with domain name generation, see the "search" command. 

For multiple domain checks, seperate domain names with comma or space

Examples:
	dmn check example.com
	dmn check example.com example.org
`,
		Run: func(cmd *cobra.Command, args []string) {
			runCheckCmd(flags, args)
		},
	}

	cmd.Flags().BoolVarP(&flags.showWhois, "whois", "w", false, "Show whois information of the domain name")
	cmd.Flags().BoolVarP(&flags.benchmark, "benchmark", "b", false, "Show benchmark information")
	cmd.Flags().IntVarP(&flags.parallel, "parallel", "p", 4, "Number of concurrent workers that check domains")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCheckCommand())
}

func runCheckCmd(flags checkCmdFlags, args []string) {
	if flags.benchmark {
		defer dev.PrintElapsedTime("Running time", time.Now())
	}
	allNames := getNames(args)

	if len(allNames) == 0 {
		fmt.Println("Error: No domain name is given")
		return
	}

	out := whois.LookupMultiple(nil, allNames, flags.parallel)

	for record := range out {
		printRecord(record, flags.benchmark, flags.showWhois)

	}

}

func getNames(args []string) []string {
	allNames := []string{}
	stdinStr, err := input.Stdin()
	if err == nil && stdinStr != "" {
		allNames = input.GrabDomainNames(stdinStr)
	} else {
		for _, arg := range args {
			names := input.GrabDomainNames(arg)
			allNames = append(allNames, names...)
		}
	}

	return allNames
}

func printRecord(record whois.Record, showBenchmark bool, showWhois bool) {
	benchmarkStr := ""
	if showBenchmark {
		benchmarkStr = fmt.Sprintf(" (%.3f)", float32(record.Elapsed/time.Millisecond)/1000)
	}

	if record.Error != nil {
		color.Red("❗ %s: error (%s)%s", record.Name, record.Error.Error(), benchmarkStr)
	} else if record.Attributes.Available {
		color.Green("✔ %s: available%s", record.Name, benchmarkStr)
	} else {
		color.Yellow("✘ %s: unavailable%s", record.Name, benchmarkStr)
	}

	if record.Error == nil && showWhois {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println(record.Response)
		fmt.Println(strings.Repeat("=", 80))
	}
}
