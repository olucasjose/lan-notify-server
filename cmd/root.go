package cmd

import (
	"fmt"
	"os"

	"lan-notify/internal/i18n"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "lan-notify",
	Version:       "1.0.0",
	Short:         i18n.T("cmd_root_short"),
	Long:          i18n.T("cmd_root_long"),
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", i18n.T("err_root_execute"), err)
		os.Exit(1)
	}
}

func init() {
	// Deixamos a opção padrão ativa para o script das dicas funcionar
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	usageTemplate := fmt.Sprintf(`%s{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [%s]{{end}}{{if .HasExample}}

%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{if eq .Name "help"}}%s{{else}}{{.Short}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

%s{{end}}
`,
		i18n.T("usage_prefix"),
		i18n.T("usage_command"),
		i18n.T("usage_examples"),
		i18n.T("usage_available_commands"),
		i18n.T("usage_help_description"),
		i18n.T("usage_flags"),
		i18n.T("usage_global_flags"),
		fmt.Sprintf(i18n.T("usage_help_footer"), i18n.T("usage_command")),
	)

	rootCmd.SetUsageTemplate(usageTemplate)
}
