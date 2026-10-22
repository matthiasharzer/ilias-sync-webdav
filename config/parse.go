package config

import (
	"net/url"
	"strings"

	"github.com/bigkevmcd/go-configparser"
)

func parseDefaultSection(p *configparser.ConfigParser) (*url.URL, string, error) {
	urlStr, err := p.Get("DEFAULT", "url")
	if err != nil {
		return nil, "", err
	}
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}

	localDir, err := p.Get("DEFAULT", "directory")
	if err != nil {
		return nil, "", err
	}

	return urlParsed, localDir, nil
}

func ParseFile(path string) (*Config, error) {
	p, err := configparser.NewConfigParserFromFile(path)
	if err != nil {
		return nil, err
	}

	var targets []Target
	sections := p.Sections()
	for _, section := range sections {
		if section == "DEFAULT" {
			continue
		}
		if !strings.HasPrefix(section, "sync:") {
			continue
		}
		name := strings.TrimPrefix(section, "sync:")
		courseID, err := p.Get(section, "id")
		if err != nil {
			return nil, err
		}
		targets = append(targets, Target{
			Name:     name,
			CourseID: courseID,
		})
	}
	webdavURL, localDir, err := parseDefaultSection(p)
	if err != nil {
		return nil, err
	}

	return &Config{
		Targets:            targets,
		ILIASWebDAVBaseURL: *webdavURL,
		LocalDirectory:     localDir,
	}, nil
}
