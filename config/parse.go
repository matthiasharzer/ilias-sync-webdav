package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/bigkevmcd/go-configparser"
)

type parser = func(p *configparser.ConfigParser, config *Config, section string) error

var parsers = map[string]parser{
	"sync": parseSyncSection,
	"auth": parseAuthSection,
}

func parseDefaultSection(p *configparser.ConfigParser, config *Config) error {
	urlStr, err := p.Get("DEFAULT", "url")
	if err != nil {
		return err
	}
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	localDir, err := p.Get("DEFAULT", "directory")
	if err != nil {
		return err
	}

	hasExistsStrategy, err := p.HasOption("DEFAULT", "file_already_exists_strategy")
	if err != nil {
		return err
	}
	if hasExistsStrategy {
		alreadyExistsStrategyStr, err := p.Get("DEFAULT", "file_already_exists_strategy")
		if err != nil {
			return err
		}
		switch alreadyExistsStrategyStr {
		case "skip":
			config.AlreadyExistsStrategy = Skip
		case "overwrite":
			config.AlreadyExistsStrategy = Overwrite
		default:
			return fmt.Errorf("invalid file_already_exists_strategy: %s", alreadyExistsStrategyStr)
		}
	} else {
		config.AlreadyExistsStrategy = Skip // default strategy
	}

	config.ILIASWebDAVBaseURL = *urlParsed
	config.LocalDirectory = localDir
	return nil
}

func parseSyncSection(p *configparser.ConfigParser, config *Config, section string) error {
	courseID, err := p.Get(section, "id")
	if err != nil {
		return err
	}
	name := strings.TrimPrefix(section, "sync:")
	config.Targets = append(config.Targets, Target{
		Name:     name,
		CourseID: courseID,
	})
	return nil
}

func parseAuthSection(p *configparser.ConfigParser, config *Config, section string) error {
	authType := strings.TrimPrefix(section, "auth:")
	switch authType {
	case "basic":
		username, err := p.Get(section, "username")
		if err != nil {
			return err
		}
		password, err := p.Get(section, "password")
		if err != nil {
			return err
		}
		config.Username = username
		config.Password = password
	case "file":
		filePath, err := p.Get(section, "path")
		if err != nil {
			return err
		}
		// Read username and password from the file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		lines := strings.Split(string(data), "\n")
		if len(lines) < 2 {
			return nil // or return an error indicating the file format is incorrect
		}
		config.Username = strings.TrimSpace(lines[0])
		config.Password = strings.TrimSpace(lines[1])
	case "env":
		username := os.Getenv("ILIAS_USERNAME")
		password := os.Getenv("ILIAS_PASSWORD")
		if username == "" || password == "" {
			return fmt.Errorf("environment variables ILIAS_USERNAME or ILIAS_PASSWORD are not set")
		}
		config.Username = username
		config.Password = password
	default:
		return fmt.Errorf("unsupported auth type: %s", authType)
	}
	return nil
}

func getParser(section string) (parser, bool) {
	for prefix, parser := range parsers {
		if strings.HasPrefix(section, prefix) {
			return parser, true
		}
	}
	return nil, false
}

func validateConfig(config *Config) error {
	if config.ILIASWebDAVBaseURL.String() == "" {
		return fmt.Errorf("ILIAS WebDAV base URL is not set")
	}
	if config.LocalDirectory == "" {
		return fmt.Errorf("local directory is not set")
	}
	if len(config.Targets) == 0 {
		return fmt.Errorf("no targets specified")
	}
	if config.Username == "" || config.Password == "" {
		return fmt.Errorf("username or password is not set")
	}
	return nil
}

func ParseFile(path string) (*Config, error) {
	p, err := configparser.NewConfigParserFromFile(path)
	if err != nil {
		return nil, err
	}

	sections := p.Sections()

	config := &Config{}

	// The "DEFAULT" section is not included in the list of sections, so we need to parse it separately
	err = parseDefaultSection(p, config)
	if err != nil {
		return nil, err
	}

	for _, section := range sections {
		sectionParser, ok := getParser(section)
		if !ok {
			continue
		}
		err := sectionParser(p, config, section)
		if err != nil {
			return nil, err
		}
	}

	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
