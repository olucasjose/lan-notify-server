package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "lan-notify",
	Version: "1.0.0",
	Short:   "LAN Notify é um sistema de notificação em rede local",
	Long: `LAN Notify é um serviço em segundo plano (daemon) e um cliente CLI (linha de comando)
para enviar e receber notificações desktop nativas através da rede local usando mDNS e HTTPS seguro.

Dicas de Autocompletar:
Para habilitar o [TAB] no terminal, gere o script correspondente ao seu shell.
Exemplo Linux (Bash):
  lan-notify completion bash | sudo tee /etc/bash_completion.d/lan-notify > /dev/null
  exec bash

Exemplo Termux (Android):
  lan-notify completion bash > /data/data/com.termux/files/usr/etc/bash_completion.d/lan-notify
  exit (Reinicie a sessão do terminal completamente)`,
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
