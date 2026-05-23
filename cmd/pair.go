package cmd

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lan-notify/internal/client"
	"lan-notify/internal/config"
	"lan-notify/internal/discovery"

	"github.com/spf13/cobra"
)

var pairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Manage mTLS pairing",
}

type PairingPinData struct {
	PIN       string    `json:"pin"`
	ExpiresAt time.Time `json:"expires_at"`
}

var pairGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a pairing PIN for this server",
	Run: func(cmd *cobra.Command, args []string) {
		appDir, err := config.GetConfigDir()
		if err != nil {
			log.Fatalf("Failed to get config dir: %v", err)
		}

		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// 1. Healthcheck to see if the background server is running
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", cfg.Port), 1*time.Second)
		if err != nil {
			fmt.Printf("⚠️ Aviso: O 'lan-notify server' não parece estar rodando neste computador (porta %d).\n", cfg.Port)
			fmt.Println("Certifique-se de iniciá-lo em outro terminal ou serviço para que o pareamento funcione.")
			fmt.Println()
		} else {
			conn.Close()
			fmt.Printf("✅ Servidor local detectado (porta %d).\n", cfg.Port)
		}

		pinNum, err := rand.Int(rand.Reader, big.NewInt(10000))
		if err != nil {
			log.Fatalf("Failed to generate random PIN: %v", err)
		}
		pinStr := fmt.Sprintf("%04d", pinNum.Int64())

		data := PairingPinData{
			PIN:       pinStr,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}

		b, err := json.Marshal(data)
		if err != nil {
			log.Fatalf("Failed to encode PIN: %v", err)
		}

		pinPath := filepath.Join(appDir, "pairing_pin.json")
		if err := os.WriteFile(pinPath, b, 0600); err != nil {
			log.Fatalf("Failed to save PIN: %v", err)
		}

		fmt.Printf("PIN gerado: %s\n", pinStr)
		fmt.Println("No outro dispositivo, execute: lan-notify pair connect @<este-dispositivo>")
		fmt.Printf("⏳ Aguardando conexão do outro dispositivo com o PIN %s...\n", pinStr)

		// 2. Polling loop: Wait for the background server to delete the file upon success
		timeout := time.After(5 * time.Minute)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				os.Remove(pinPath)
				fmt.Println("\n❌ O tempo limite de 5 minutos expirou e o PIN foi invalidado.")
				os.Exit(1)
			case <-ticker.C:
				if _, err := os.Stat(pinPath); os.IsNotExist(err) {
					fmt.Println("\n✅ Pareamento concluído com sucesso pelo servidor em background!")
					return
				}
			}
		}
	},
}

var pairConnectCmd = &cobra.Command{
	Use:   "connect <target> [pin]",
	Short: "Connect to a target server using a pairing PIN",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		target := strings.TrimPrefix(args[0], "@")
		var pin string

		if len(args) > 1 {
			pin = args[1]
		} else {
			fmt.Printf("Digite o PIN fornecido pelo @%s: ", target)
			reader := bufio.NewReader(os.Stdin)
			pin, _ = reader.ReadString('\n')
			pin = strings.TrimSpace(pin)
		}

		if pin == "" {
			log.Fatal("PIN is required")
		}

		var resolvedIP string
		var resolvedPort int

		if net.ParseIP(target) != nil {
			resolvedIP = target
			resolvedPort = cfg.Port
		} else {
			disc := discovery.New()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			resolvedIP, resolvedPort, err = disc.ResolveTarget(ctx, target)
			if err != nil {
				log.Fatalf("Target not found on network")
			}
		}

		fmt.Println("Conectando ao servidor...")
		fingerprint, err := client.ConnectPair(cfg, resolvedIP, resolvedPort, target, pin)
		if err != nil {
			log.Fatalf("Falha no pareamento: %v", err)
		}

		// Save the fingerprint
		cfg.PinnedPeers[target] = fingerprint
		if err := cfg.Save(); err != nil {
			log.Fatalf("Falha ao salvar configuração: %v", err)
		}

		fmt.Printf("Pareamento concluído com sucesso com @%s!\n", target)
	},
}

func init() {
	pairCmd.AddCommand(pairGenerateCmd)
	pairCmd.AddCommand(pairConnectCmd)
	rootCmd.AddCommand(pairCmd)
}
