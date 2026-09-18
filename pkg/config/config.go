package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/auroq/botropolis/pkg/proto"
)

const (
	envPrefix  = "BOTROPOLIS"
	appDir     = "botropolis"
	configName = "config"

	KeyHome        = "home"
	KeySocket      = "socket"
	KeyTerminal    = "terminal"
	KeyHookCommand = "hook_command"
	KeyParkedDays  = "parked_days"
	KeyCodexHome   = "codex_home"
	KeyProjection  = "projection"
	KeyRenderScale = "render_scale"

	DefaultParkedDays = 7
)

const DefaultProjection = "iso"

type Config struct {
	Home        string
	HomeSet     bool
	Socket      string
	SocketSet   bool
	Terminal    string
	HookCommand string
	ParkedDays  int
	CodexHome   string
	Projection  string
	RenderScale float64
	File        string
}

func NewViper() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetDefault(KeyHome, "")
	v.SetDefault(KeySocket, "")
	v.SetDefault(KeyTerminal, "")
	v.SetDefault(KeyHookCommand, "botropolis-hook")
	v.SetDefault(KeyParkedDays, DefaultParkedDays)
	v.SetDefault(KeyCodexHome, "")
	v.SetDefault(KeyProjection, DefaultProjection)
	v.SetDefault(KeyRenderScale, 0.0)
	v.SetConfigName(configName)
	v.AddConfigPath(filepath.Join(configHome(), appDir))
	return v
}

func BindFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.String(KeyHome, "", "home directory holding .claude (default: $HOME)")
	flags.String(KeySocket, "", "daemon socket (default: $XDG_RUNTIME_DIR/botropolis/botropolis.sock)")
	flags.Int(KeyParkedDays, DefaultParkedDays, "how many days of parked sessions to catalogue (0 disables)")
	flags.String(KeyProjection, DefaultProjection, "how the city is drawn: iso or top")
	flags.Float64(KeyRenderScale, 0, "chrome and pixel scale (default: follow the display)")
	_ = v.BindPFlag(KeyHome, flags.Lookup(KeyHome))
	_ = v.BindPFlag(KeySocket, flags.Lookup(KeySocket))
	_ = v.BindPFlag(KeyParkedDays, flags.Lookup(KeyParkedDays))
	_ = v.BindPFlag(KeyProjection, flags.Lookup(KeyProjection))
	_ = v.BindPFlag(KeyRenderScale, flags.Lookup(KeyRenderScale))
}

func New(v *viper.Viper) (*Config, error) {
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
	}
	cfg := &Config{
		Home:        v.GetString(KeyHome),
		HomeSet:     v.GetString(KeyHome) != "",
		Socket:      v.GetString(KeySocket),
		SocketSet:   v.GetString(KeySocket) != "",
		Terminal:    v.GetString(KeyTerminal),
		HookCommand: v.GetString(KeyHookCommand),
		ParkedDays:  v.GetInt(KeyParkedDays),
		CodexHome:   v.GetString(KeyCodexHome),
		Projection:  v.GetString(KeyProjection),
		RenderScale: v.GetFloat64(KeyRenderScale),
		File:        v.ConfigFileUsed(),
	}
	if cfg.Home == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		cfg.Home = home
	}
	if cfg.Socket == "" {
		cfg.Socket = proto.SocketPath()
	}
	if cfg.CodexHome == "" {
		cfg.CodexHome = filepath.Join(cfg.Home, ".codex")
	}
	return cfg, nil
}

func configHome() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config")
}
