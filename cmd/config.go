package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"lan-notify/internal/config"
	"lan-notify/internal/i18n"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   i18n.T("cmd_config_use"),
	Short: i18n.T("cmd_config_short"),
	Run: func(cmd *cobra.Command, args []string) {
		appDir, err := config.GetConfigDir()
		if err != nil {
			log.Fatalf("%s: %v", i18n.T("err_config_dir"), err)
		}

		path := filepath.Join(appDir, "config.json")

		// Inform if config already exists
		if _, err := os.Stat(path); err == nil {
			fmt.Print(i18n.T("warn_config_exists", path))
		}

		reader := bufio.NewReader(os.Stdin)

		// Get Device Name
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "my-device"
		}
		fmt.Print(i18n.T("prompt_device_name", hostname))
		deviceName, _ := reader.ReadString('\n')
		deviceName = strings.TrimSpace(deviceName)
		if deviceName == "" {
			deviceName = hostname
		}

		// Save Configuration
		if err := os.MkdirAll(appDir, 0755); err != nil {
			log.Fatalf("%s: %v", i18n.T("err_create_dir"), err)
		}

		cfg := config.Config{
			DeviceName:  deviceName,
			Port:        42931,
			PinnedPeers: make(map[string]string),
		}

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("%s: %v", i18n.T("err_create_file"), err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(cfg); err != nil {
			log.Fatalf("%s: %v", i18n.T("err_save_config"), err)
		}

		fmt.Println(i18n.T("msg_config_saved_success"))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
