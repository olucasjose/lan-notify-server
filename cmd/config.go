package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"lan-notify/internal/config"

	"github.com/spf13/cobra"
)

// generateRandomToken creates a secure 16-byte hex token.
func generateRandomToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback_token_123"
	}
	return hex.EncodeToString(bytes)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configures the lan-notify application interactively",
	Run: func(cmd *cobra.Command, args []string) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatalf("Failed to determine config directory: %v", err)
		}

		appDir := filepath.Join(configDir, "lan-notify")
		path := filepath.Join(appDir, "config.json")

		// Inform if config already exists
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Aviso: Já existe uma configuração em %s. Ela será sobrescrita.\n", path)
		}

		reader := bufio.NewReader(os.Stdin)

		// Get Device Name
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "my-device"
		}
		fmt.Printf("Qual o nome deste dispositivo? [%s]: ", hostname)
		deviceName, _ := reader.ReadString('\n')
		deviceName = strings.TrimSpace(deviceName)
		if deviceName == "" {
			deviceName = hostname
		}

		// Get Auth Token
		fmt.Printf("Digite uma senha de segurança (ou deixe em branco para gerar uma aleatória): ")
		authToken, _ := reader.ReadString('\n')
		authToken = strings.TrimSpace(authToken)
		if authToken == "" {
			authToken = generateRandomToken()
			fmt.Printf("🔑 Senha aleatória gerada: %s\n", authToken)
		}

		// Save Configuration
		if err := os.MkdirAll(appDir, 0755); err != nil {
			log.Fatalf("Falha ao criar diretório de configuração: %v", err)
		}

		cfg := config.Config{
			DeviceName: deviceName,
			Port:       8080,
			AuthToken:  authToken,
		}

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("Falha ao criar arquivo de configuração: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(cfg); err != nil {
			log.Fatalf("Falha ao salvar configuração: %v", err)
		}

		fmt.Println("\n✅ Configuração salva com sucesso! O lan-notify já está pronto para uso.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
