package util

// Config is required for initializing the service
type Config struct {
	NodeID                     string `config:"id,required"`
	BindAddr                   string `config:"bind,required"`
	ListenAddr                 string `config:"listen,required"`
	Dir                        string `config:"dir"`
	JoinAddr                   string `config:"join"`
	DefaultWaitWindow          uint64 `config:"wait_window"`
	DefaultWaitWindowThreshold uint64 `config:"wait_window_threshold"`
	DefaultMaxWaitWindow       uint64 `config:"max_wait_window"`
	Version                    string `config:"version"`
	Commit                     string `config:"commit"`
	Date                       string `config:"date"`
	DisablePostHook            bool   `config:"dev"`
}
