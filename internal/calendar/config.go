package calendar

import (
	"fmt"
	"os"
	"strings"

	"simplified-dashboard/internal/appenv"
)

type Config struct {
	CredentialsPath string
	TokenPath       string
	Calendars       []CalendarConfig
}

type CalendarConfig struct {
	ID    string
	Label string
	Color string
}

func LoadConfigFromEnv() (Config, error) {
	config := Config{
		CredentialsPath: os.Getenv("GOOGLE_CALENDAR_CREDENTIALS_PATH"),
		TokenPath:       os.Getenv("GOOGLE_CALENDAR_TOKEN_PATH"),
		Calendars:       calendarsFromEnv(),
	}

	config.applyDefaults()
	if err := config.validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (config *Config) applyDefaults() {
	config.CredentialsPath = appenv.ExpandPath(config.CredentialsPath)
	config.TokenPath = appenv.ExpandPath(config.TokenPath)

	for i := range config.Calendars {
		config.Calendars[i].ID = strings.TrimSpace(config.Calendars[i].ID)
		config.Calendars[i].Label = strings.TrimSpace(config.Calendars[i].Label)
		config.Calendars[i].Color = strings.TrimSpace(config.Calendars[i].Color)
	}
}

func (config Config) validate() error {
	if config.CredentialsPath == "" {
		return fmt.Errorf("GOOGLE_CALENDAR_CREDENTIALS_PATH is required")
	}
	if config.TokenPath == "" {
		return fmt.Errorf("GOOGLE_CALENDAR_TOKEN_PATH is required")
	}
	if len(config.Calendars) == 0 {
		return fmt.Errorf("GOOGLE_CALENDAR_IDS must include at least one calendar")
	}
	for index, calendar := range config.Calendars {
		if calendar.ID == "" {
			return fmt.Errorf("calendar config calendar %d must include an id", index+1)
		}
	}
	return nil
}

func (calendar CalendarConfig) DisplayLabel() string {
	if calendar.Label != "" {
		return calendar.Label
	}
	return calendar.ID
}

func calendarsFromEnv() []CalendarConfig {
	ids := splitEnvList(os.Getenv("GOOGLE_CALENDAR_IDS"))
	labels := splitEnvList(os.Getenv("GOOGLE_CALENDAR_LABELS"))
	colors := splitEnvList(os.Getenv("GOOGLE_CALENDAR_COLORS"))

	calendars := make([]CalendarConfig, 0, len(ids))
	for i, id := range ids {
		calendar := CalendarConfig{ID: id}
		if i < len(labels) {
			calendar.Label = labels[i]
		}
		if i < len(colors) {
			calendar.Color = colors[i]
		}
		calendars = append(calendars, calendar)
	}
	return calendars
}

func splitEnvList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}
