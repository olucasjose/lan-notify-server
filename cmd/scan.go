package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"lan-notify/internal/discovery"
	"lan-notify/internal/i18n"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   i18n.T("cmd_scan_use"),
	Short: i18n.T("cmd_scan_short"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(i18n.T("msg_scanning"))

		disc := discovery.New()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		devices, err := disc.Scan(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "failed to join") || strings.Contains(err.Error(), "operation not permitted") {
				fmt.Println(i18n.T("err_network_block"))
				fmt.Println(i18n.T("tip_network_block"))
				os.Exit(1)
			}
			log.Fatalf("%s: %v", i18n.T("err_scan_fail"), err)
		}

		if len(devices) == 0 {
			fmt.Println(i18n.T("msg_no_devices"))
			return
		}

		fmt.Println(i18n.T("msg_devices_found"))
		for _, dev := range devices {
			fmt.Printf("   - @%s\n", dev)
		}
		fmt.Println(i18n.T("tip_send_msg"))
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
