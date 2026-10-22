package config

import "net/url"

type Target struct {
	Name     string
	CourseID string
}

type Config struct {
	Targets            []Target
	ILIASWebDAVBaseURL url.URL
	LocalDirectory     string
}
