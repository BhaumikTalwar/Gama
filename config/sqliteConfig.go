package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SqliteConfig struct {
	Path         string `mapstructure:"path" yaml:"path" validate:"required"`
	JournalMode  string `mapstructure:"journal_mode" yaml:"journal_mode"`
	Synchronous  string `mapstructure:"synchronous" yaml:"synchronous"`
	TempStore    string `mapstructure:"temp_store" yaml:"temp_store"`
	AutoVacuum   string `mapstructure:"auto_vacuum" yaml:"auto_vacuum"`
	CacheSize    int    `mapstructure:"cache_size" yaml:"cache_size"`
	BusyTimeout  int    `mapstructure:"busy_timeout" yaml:"busy_timeout"`
	MmapSize     int    `mapstructure:"mmap_size" yaml:"mmap_size"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	ForeignKeys  bool   `mapstructure:"foreign_keys" yaml:"foreign_keys"`
}

func setSqliteDefaults(prefix string) {
	viper.SetDefault(prefix+".path", "./data.db")
	viper.SetDefault(prefix+".journal_mode", "WAL")
	viper.SetDefault(prefix+".synchronous", "NORMAL")
	viper.SetDefault(prefix+".busy_timeout", 5000)
	viper.SetDefault(prefix+".cache_size", -2000)
	viper.SetDefault(prefix+".foreign_keys", true)
	viper.SetDefault(prefix+".max_open_conns", 1)
	viper.SetDefault(prefix+".max_idle_conns", 1)
}

func registerSqliteFlags(cmd *cobra.Command, prefix string) {
	pathFlag := fmt.Sprintf("%s-path", prefix)
	journalFlag := fmt.Sprintf("%s-journal-mode", prefix)
	syncFlag := fmt.Sprintf("%s-synchronous", prefix)
	tempStoreFlag := fmt.Sprintf("%s-temp-store", prefix)
	vacuumFlag := fmt.Sprintf("%s-auto-vacuum", prefix)

	cacheFlag := fmt.Sprintf("%s-cache-size", prefix)
	busyFlag := fmt.Sprintf("%s-busy-timeout", prefix)
	mmapFlag := fmt.Sprintf("%s-mmap-size", prefix)

	maxOpenFlag := fmt.Sprintf("%s-max-open-conns", prefix)
	maxIdleFlag := fmt.Sprintf("%s-max-idle-conns", prefix)
	fkFlag := fmt.Sprintf("%s-foreign-keys", prefix)

	cmd.Flags().String(pathFlag, "", "Path to SQLite database file")
	cmd.Flags().String(journalFlag, "", "Journal mode (DELETE, TRUNCATE, WAL)")
	cmd.Flags().String(syncFlag, "", "Synchronous mode (OFF, NORMAL, FULL)")
	cmd.Flags().String(tempStoreFlag, "", "Temp store (DEFAULT, FILE, MEMORY)")
	cmd.Flags().String(vacuumFlag, "", "Auto vacuum mode")

	cmd.Flags().Int(cacheFlag, 0, "Cache size (positive=pages, negative=kibibytes)")
	cmd.Flags().Int(busyFlag, 0, "Busy timeout in milliseconds")
	cmd.Flags().Int(mmapFlag, 0, "Memory map size")

	cmd.Flags().Int(maxOpenFlag, 0, "Max open connections")
	cmd.Flags().Int(maxIdleFlag, 0, "Max idle connections")
	cmd.Flags().Bool(fkFlag, false, "Enable Foreign Keys")

	_ = viper.BindPFlag(prefix+".path", cmd.Flags().Lookup(pathFlag))
	_ = viper.BindPFlag(prefix+".journal_mode", cmd.Flags().Lookup(journalFlag))
	_ = viper.BindPFlag(prefix+".synchronous", cmd.Flags().Lookup(syncFlag))
	_ = viper.BindPFlag(prefix+".temp_store", cmd.Flags().Lookup(tempStoreFlag))
	_ = viper.BindPFlag(prefix+".auto_vacuum", cmd.Flags().Lookup(vacuumFlag))

	_ = viper.BindPFlag(prefix+".cache_size", cmd.Flags().Lookup(cacheFlag))
	_ = viper.BindPFlag(prefix+".busy_timeout", cmd.Flags().Lookup(busyFlag))
	_ = viper.BindPFlag(prefix+".mmap_size", cmd.Flags().Lookup(mmapFlag))

	_ = viper.BindPFlag(prefix+".max_open_conns", cmd.Flags().Lookup(maxOpenFlag))
	_ = viper.BindPFlag(prefix+".max_idle_conns", cmd.Flags().Lookup(maxIdleFlag))
	_ = viper.BindPFlag(prefix+".foreign_keys", cmd.Flags().Lookup(fkFlag))
}

func (c *SqliteConfig) validate() error {
	// no-op
	return nil
}
