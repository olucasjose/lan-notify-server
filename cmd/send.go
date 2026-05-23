package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"lan-notify/internal/client"
	"lan-notify/internal/config"
	"lan-notify/internal/discovery"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [message] [@target]",
	Short: "Sends a notification to a target device",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Configuration
		cfg, err := config.Load()
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("❌ Configuração não encontrada em ~/.config/lan-notify.\n💡 Dica: Rode o comando 'lan-notify config' para configurar seu dispositivo pela primeira vez.")
			}
			log.Fatalf("Failed to load config: %v", err)
		}

		var target string
		var messageParts []string

		// 2. Parse Arguments
		for _, arg := range args {
			if strings.HasPrefix(arg, "@") {
				target = strings.TrimPrefix(arg, "@")
			} else {
				messageParts = append(messageParts, arg)
			}
		}

		message := strings.Join(messageParts, " ")
		if message == "" {
			log.Fatalf("No message provided")
		}

		if target == "" {
			log.Fatalf("No target provided (use @target)")
		}

		var resolvedIP string
		var resolvedPort int

		// 3. Resolve Target (Direct IP or mDNS)
		if net.ParseIP(target) != nil {
			resolvedIP = target
			resolvedPort = cfg.Port
		} else {
			disc := discovery.New()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			var err error
			resolvedIP, resolvedPort, err = disc.ResolveTarget(ctx, target)
			if err != nil {
				log.Fatalf("Target '%s' not found on local network: %v", target, err)
			}
		}

		// 4. Determine Token to use
		token := cfg.AuthToken // default to own token
		if savedToken, ok := cfg.KnownPeers[target]; ok && savedToken != "" {
			token = savedToken
		}

		req := client.NotificationRequest{
			Title:   "Message from " + cfg.DeviceName,
			Message: message,
			Urgency: "normal",
		}

		// 5. Try Sending
		err = client.SendNotification(resolvedIP, resolvedPort, token, req)

		// If 401 Unauthorized, prompt for password
		if err != nil && strings.Contains(err.Error(), "401") {
			fmt.Printf("🔒 O dispositivo @%s exige autenticação.\n", target)
			fmt.Printf("Digite a senha do %s para se conectar: ", target)

			reader := bufio.NewReader(os.Stdin)
			newToken, _ := reader.ReadString('\n')
			newToken = strings.TrimSpace(newToken)

			// Try again with new token
			err = client.SendNotification(resolvedIP, resolvedPort, newToken, req)
			if err == nil {
				// Save it for future uses!
				cfg.KnownPeers[target] = newToken
				if saveErr := cfg.Save(); saveErr != nil {
					fmt.Printf("Aviso: Falha ao salvar a senha para uso futuro: %v\n", saveErr)
				} else {
					fmt.Println("🔑 Senha salva com sucesso!")
				}
			}
		}

		if err != nil {
			log.Fatalf("Falha ao enviar notificação: %v", err)
		}

		log.Println("✅ Notificação entregue com sucesso!")
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
