package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"lan-notify/internal/discovery"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans the local network for available lan-notify devices",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Escaneando a rede local por dispositivos (aguarde 3 segundos)...")

		disc := discovery.New()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		devices, err := disc.Scan(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "failed to join") || strings.Contains(err.Error(), "operation not permitted") {
				fmt.Println("\n❌ Erro: Bloqueio de rede detectado (provavelmente pelo Android).")
				fmt.Println("💡 Dica: O sistema operacional bloqueia varreduras Multicast (mDNS). O comando 'scan' não funciona neste ambiente.")
				fmt.Println("         Para enviar mensagens, forneça o IP do computador alvo diretamente (ex: lan-notify send \"msg\" 192.168.0.10).")
				os.Exit(1)
			}
			log.Fatalf("\n❌ Falha ao escanear a rede: %v", err)
		}

		if len(devices) == 0 {
			fmt.Println("Nenhum dispositivo encontrado na rede local.")
			return
		}

		fmt.Println("\n🖥️  Dispositivos encontrados:")
		for _, dev := range devices {
			fmt.Printf("   - @%s\n", dev)
		}
		fmt.Println("\n💡 Para enviar uma notificação, use: lan-notify send \"Sua mensagem\" @<nome-do-dispositivo>")
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
