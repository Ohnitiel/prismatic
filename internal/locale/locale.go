package locale

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type CliFlags struct {
	Config        string `toml:"config"`
	Environment   string `toml:"environment"`
	Connections   string `toml:"connections"`
	OutputFormat  string `toml:"output_format"`
	NoCache       string `toml:"no_cache"`
	NoSingleSheet string `toml:"no_single_sheet"`
	NoSingleFile  string `toml:"no_single_file"`
	Commit        string `toml:"commit"`
}

type CliCommands struct {
	Export string `toml:"export"`
	Run    string `toml:"run"`
	Check  string `toml:"check"`
}

type CliArgs struct {
	Export string `toml:"export"`
	Run    string `toml:"run"`
}

type CliSection struct {
	Description string      `toml:"description"`
	Flags       CliFlags    `toml:"flags"`
	Commands    CliCommands `toml:"commands"`
	Args        CliArgs     `toml:"args"`
}

type ErrorsSection struct {
	InvalidEnvironment  string `toml:"invalid_environment"`
	OutputFormatNotImpl string `toml:"output_format_not_implemented"`
	OutputFormatEmpty   string `toml:"output_format_empty"`
	ConnectionFailed    string `toml:"connection_failed"`
	QueryFailed         string `toml:"query_failed"`
	ContextDeadline     string `toml:"context_deadline"`
	NoDataReturned      string `toml:"no_data_returned"`
}

type Locale struct {
	CLI    CliSection    `toml:"cli"`
	Errors ErrorsSection `toml:"errors"`
}

func DetectSystemLocale() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		return "en_US"
	}

	cleanLang := strings.Split(lang, ".")[0]

	return strings.ReplaceAll(cleanLang, "-", "_")
}

func Load(localeName string) (*Locale, error) {
	if localeName == "" || strings.ToLower(localeName) == "auto" {
		localeName = DetectSystemLocale()
	}

	localePath := filepath.Join("config", "locales", fmt.Sprintf("%s.toml", localeName))

	if _, err := os.Stat(localePath); os.IsNotExist(err) {
		localePath = filepath.Join("config", "locales", "en_US.toml")
	}

	var l Locale
	if _, err := toml.DecodeFile(localePath, &l); err != nil {
		return nil, fmt.Errorf("failed to load locale file %s: %w", localePath, err)
	}

	return &l, nil
}
