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
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Deixamos a opção padrão ativa para o script das dicas funcionar
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SetUsageTemplate(`Uso:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [comando]{{end}}{{if .HasExample}}

Exemplos:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Comandos Disponíveis:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{if eq .Name "help"}}Exibe informações de ajuda para os comandos{{else}}{{.Short}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Flags Globais:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [comando] --help" para mais informações sobre um comando específico.{{end}}
`)
}
