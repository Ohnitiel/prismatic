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

type ExitMessages struct {
	Success     string `toml:"success"`
	PartialFail string `toml:"partial_fail"`
	FullFail    string `toml:"full_fail"`
}

type Locale struct {
	CLI    CliSection    `toml:"cli"`
	Errors ErrorsSection `toml:"errors"`
	Logs   LogsSection   `toml:"logs"`
	ExitMessages   ExitMessages  `toml:"exit_messages"`
}

type LogsSection struct {
	CommittingTransaction      string `toml:"committing_transaction"`
	ConnectionFailed           string `toml:"connection_failed"`
	ContextAlreadyCancelled    string `toml:"context_already_cancelled"`
	ErrorClosingFile           string `toml:"error_closing_file"`
	ErrorFlushingData          string `toml:"error_flushing_data"`
	ErrorIdentifyingColumns    string `toml:"error_identifying_columns"`
	ErrorPreparingStatement    string `toml:"error_preparing_statement"`
	ErrorRunningQuery          string `toml:"error_running_query"`
	ErrorRunningQueryOnConn    string `toml:"error_running_query_on_conn"`
	ErrorSavingFile            string `toml:"error_saving_file"`
	ErrorScanningRows          string `toml:"error_scanning_rows"`
	ErrorStartingTransaction   string `toml:"error_starting_transaction"`
	ErrorWritingData           string `toml:"error_writing_data"`
	GenericRowError            string `toml:"generic_row_error"`
	IdentifiedQueryType        string `toml:"identified_query_type"`
	NoHostSpecified            string `toml:"no_host_specified"`
	QueryResultCache           string `toml:"query_result_cache"`
	QuerySuccessfulOnConn      string `toml:"query_successful_on_conn"`
	RollingBackTransaction     string `toml:"rolling_back_transaction"`
	RunningSelectWithoutSaving string `toml:"running_select_without_saving"`
	RunningQueryOnConn         string `toml:"running_query_on_conn"`
	SkippingConnectionError    string `toml:"skipping_connection_error"`
	UnableIdentifyQueryType    string `toml:"unable_identify_query_type"`
	CacheEntryExpired          string `toml:"cache_entry_expired"`
	EnvDisabled                string `toml:"env_disabled"`
	QuerySummary               string `toml:"query_summary"`
}

var L *Locale

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
