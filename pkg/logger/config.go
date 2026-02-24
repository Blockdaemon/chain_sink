package logger

type Config struct {
	Level             string `mapstructure:"level" default:"info" validate:"oneof=debug info warn error fatal panic"`
	DisableStacktrace bool   `mapstructure:"disable_stack_trace"`
	Development       bool   `mapstructure:"dev" default:"false"`
}
