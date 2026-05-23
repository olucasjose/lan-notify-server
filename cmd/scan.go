package cmd

import (
	"context"
	"fmt"
	"log"
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
			log.Fatalf("Falha ao escanear a rede: %v", err)
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
