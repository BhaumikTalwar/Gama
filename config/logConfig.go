package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type LoggingConfig struct {
	Level    string `mapstructure:"level" yaml:"level"`
	Format   string `mapstructure:"format" yaml:"format"`
	Output   string `mapstructure:"output" yaml:"output"`
	FilePath string `mapstructure:"file_path" yaml:"file_path"`

	MaxSizeMB  int  `mapstructure:"max_size_mb" yaml:"max_size_mb"`
	MaxBackups int  `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAgeDays int  `mapstructure:"max_age_days" yaml:"max_age_days"`
	Compress   bool `mapstructure:"compress" yaml:"compress"`
	AddSource  bool `mapstructure:"add_source" yaml:"add_source"`
}

func setLogDefaults(prefix string) {
	viper.SetDefault(prefix+".level", "debug")
	viper.SetDefault(prefix+".format", "json")
	viper.SetDefault(prefix+".output", "stdout")

	viper.SetDefault(prefix+".file_path", "./logs/app.log")

	viper.SetDefault(prefix+".max_size_mb", 100)
	viper.SetDefault(prefix+".max_backups", 5)
	viper.SetDefault(prefix+".max_age_days", 30)
	viper.SetDefault(prefix+".compress", true)
	viper.SetDefault(prefix+".add_source", true)
}

func registerLogFlags(cmd *cobra.Command, prefix string) {
	levelFlag := fmt.Sprintf("%s-level", prefix)
	formatFlag := fmt.Sprintf("%s-format", prefix)
	outputFlag := fmt.Sprintf("%s-output", prefix)
	fileFlag := fmt.Sprintf("%s-file", prefix)

	maxSizeFlag := fmt.Sprintf("%s-max-size-mb", prefix)
	maxBackupsFlag := fmt.Sprintf("%s-max-backups", prefix)
	maxAgeFlag := fmt.Sprintf("%s-max-age-days", prefix)
	compressFlag := fmt.Sprintf("%s-compress", prefix)
	sourceFlag := fmt.Sprintf("%s-add-aource", prefix)

	cmd.Flags().String(levelFlag, "", "Log level (debug, info, warn, error)")
	cmd.Flags().String(formatFlag, "", "Log format (json, text)")
	cmd.Flags().String(outputFlag, "", "Log output (stdout)")
	cmd.Flags().String(fileFlag, "", "Log file path (enables file logging)")

	cmd.Flags().Int(maxSizeFlag, 0, "Max log file size in MB before rotation")
	cmd.Flags().Int(maxBackupsFlag, 0, "Max number of rotated log files")
	cmd.Flags().Int(maxAgeFlag, 0, "Max age of log files in days")
	cmd.Flags().Bool(compressFlag, false, "Compress rotated log files")
	cmd.Flags().Bool(sourceFlag, true, "Add Source info to the log")

	_ = viper.BindPFlag(prefix+".level", cmd.Flags().Lookup(levelFlag))
	_ = viper.BindPFlag(prefix+".format", cmd.Flags().Lookup(formatFlag))
	_ = viper.BindPFlag(prefix+".output", cmd.Flags().Lookup(outputFlag))
	_ = viper.BindPFlag(prefix+".file_path", cmd.Flags().Lookup(fileFlag))

	_ = viper.BindPFlag(prefix+".max_size_mb", cmd.Flags().Lookup(maxSizeFlag))
	_ = viper.BindPFlag(prefix+".max_backups", cmd.Flags().Lookup(maxBackupsFlag))
	_ = viper.BindPFlag(prefix+".max_age_days", cmd.Flags().Lookup(maxAgeFlag))
	_ = viper.BindPFlag(prefix+".compress", cmd.Flags().Lookup(compressFlag))
	_ = viper.BindPFlag(prefix+".add_source", cmd.Flags().Lookup(sourceFlag))
}

func (c *LoggingConfig) validate() error {
	level := strings.ToLower(c.Level)
	if level != "debug" && level != "info" && level != "warn" && level != "error" {
		return fmt.Errorf("Invalid Logging Level: %s", level)
	}

	if c.Format != "json" && c.Format != "text" {
		return fmt.Errorf("Invalid Log Format String: %s", c.Format)
	}

	if c.Output != "stdout" && c.Output != "file" {
		return fmt.Errorf("Invalid Log Output: %s", c.Output)
	}

	return nil
}
