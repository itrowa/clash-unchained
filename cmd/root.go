package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/itrowa/clash-unchained/internal/config"
	"github.com/itrowa/clash-unchained/internal/generator"
)

var (
	cfgFile     string
	outputFile  string
	reconfigure bool
)

var rootCmd = &cobra.Command{
	Use:   "clash-unchained",
	Short: "Turn any Clash subscription into an AI-unlocking proxy with one script",
	Long:  `clash-unchained generates a Clash Verge script that adds a chain proxy routing your AI traffic through a static long-term residential IP.`,
	RunE:  run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file path (default: ./config.yaml)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path (default: stdout)")
	rootCmd.Flags().BoolVarP(&reconfigure, "reconfigure", "r", false, "re-run the setup wizard even if config.yaml exists")
	rootCmd.Flags().BoolP("version", "v", false, "show version")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/clash-unchained")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}
}

func run(cmd *cobra.Command, args []string) error {
	if v, _ := cmd.Flags().GetBool("version"); v {
		fmt.Println("clash-unchained", version)
		return nil
	}

	// ── Wizard mode ────────────────────────────────────────────────────────
	// Trigger when: no config file found, OR --reconfigure flag is set.
	configFound := viper.ConfigFileUsed() != ""
	if !configFound || reconfigure {
		if reconfigure && configFound {
			fmt.Println("重新配置模式：将覆盖现有 config.yaml")
		}
		return RunWizard(outputFile)
	}

	// ── Normal mode (config.yaml exists) ──────────────────────────────────
	cfg, err := config.DecodeViper(viper.AllSettings())
	if err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	output, err := generator.Generate(cfg)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Printf("Generated: %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	return nil
}
