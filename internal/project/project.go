package project

import "time"

type Project struct {
	Path             string    `json:"path"`
	Name             string    `json:"name"`
	GitBranch        string    `json:"git_branch"`
	HasDockerCompose bool      `json:"has_docker_compose"`
	ProjectType      string    `json:"project_type"`
	LastModified     time.Time `json:"last_modified"`
}
