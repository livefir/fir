package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// app
	Env              string `json:"env" envconfig:"env" default:"production"`
	Name             string `json:"name" envconfig:"name" default:"goliveview-starter"`
	Domain           string `json:"domain" envconfig:"domain" default:"https://goliveview-starter.xyz"`
	Port             int    `json:"port" envconfig:"port" default:"8080"`
	HealthPath       string `json:"health_path" envconfig:"health_path" default:"/healthz"`
	ReadTimeoutSecs  int    `json:"read_timeout_secs" envconfig:"read_timeout_secs" default:"5"`
	WriteTimeoutSecs int    `json:"write_timeout_secs" envconfig:"write_timeout_secs" default:"10"`
	LogLevel         string `json:"log_level" envconfig:"log_level" default:"error"`
	LogFormatJSON    bool   `json:"log_format_json" envconfig:"log_format_json" default:"false"`
	SessionSecret    string `json:"session_secret" envconfig:"session_secret" default:"mysessionsecret"`
	APIMasterSecret  string `json:"api_master_secret" envconfig:"api_master_secret" default:"supersecretkeyyoushouldnotcommit"`

	// datasource
	Driver     string `json:"driver" envconfig:"driver" default:"sqlite3"`
	DataSource string `json:"datasource" envconfig:"datasource" default:"file:goliveview-starter.db?mode=memory&cache=shared&_fk=1"`

	// smtp
	SMTPHost       string `json:"smtp_host" envconfig:"smtp_host" default:"0.0.0.0"`
	SMTPPort       int    `json:"smtp_port,omitempty" envconfig:"smtp_port" default:"1025"`
	SMTPUser       string `json:"smtp_user" envconfig:"smtp_user" default:"myuser" `
	SMTPPass       string `json:"smtp_pass,omitempty" envconfig:"smtp_pass" default:"mypass"`
	SMTPAdminEmail string `json:"smtp_admin_email" envconfig:"smtp_admin_email" default:"noreply@goliveview-starter.xyz"`
	SMTPDebug      bool   `json:"smtp_debug" envconfig:"smtp_debug" default:"true"`

	// goth
	GoogleClientID string `json:"google_client_id" envconfig:"google_client_id"`
	GoogleSecret   string `json:"google_secret" envconfig:"google_secret"`
}

func Load(configFile string, envPrefix string) (Config, error) {
	var config Config
	if err := loadEnvironment(configFile); err != nil {
		return config, err
	}

	if err := envconfig.Process(envPrefix, &config); err != nil {
		return config, err
	}

	return config, nil
}

func loadEnvironment(filename string) error {
	var err error
	if filename != "" {
		err = godotenv.Load(filename)
		if os.IsNotExist(err) {
			return nil
		}
	} else {
		err = godotenv.Load()
		// handle if .env file does not exist, this is OK
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}
