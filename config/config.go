package config

import "net/url"

type AlreadyExistsStrategy string

const (
	Skip      AlreadyExistsStrategy = "skip"
	Overwrite AlreadyExistsStrategy = "overwrite"
)

type Target struct {
	Name     string
	CourseID string
}

type Config struct {
	Targets               []Target
	ILIASWebDAVBaseURL    url.URL
	LocalDirectory        string
	Username              string
	Password              string
	AlreadyExistsStrategy AlreadyExistsStrategy
}
