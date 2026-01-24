package metadata

import (
    "os"
    "sync"

    "gopkg.in/yaml.v3"
)

type Repo struct {
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    Command     string `yaml:"command" json:"command"`
}

type Metadata struct {
    SystemPrompt string `yaml:"system_prompt" json:"system_prompt"`
    Repos        []Repo `yaml:"repos" json:"repos"`
}

var (
    data     Metadata
    filePath string
    mu       sync.RWMutex
)

func Load(path string) error {
    mu.Lock()
    defer mu.Unlock()

    filePath = path

    file, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        data = Metadata{
            SystemPrompt: defaultSystemPrompt,
            Repos:        []Repo{},
        }
        return Save()
    }
    if err != nil {
        return err
    }

    return yaml.Unmarshal(file, &data)
}

func Save() error {
    file, err := yaml.Marshal(&data)
    if err != nil {
        return err
    }
    return os.WriteFile(filePath, file, 0644)
}

func Get() Metadata {
    mu.RLock()
    defer mu.RUnlock()
    return data
}

func AddRepo(repo Repo) error {
    mu.Lock()
    defer mu.Unlock()

    for i, r := range data.Repos {
        if r.Name == repo.Name {
            data.Repos[i] = repo
            return Save()
        }
    }

    data.Repos = append(data.Repos, repo)
    return Save()
}

func RemoveRepo(name string) bool {
    mu.Lock()
    defer mu.Unlock()

    for i, r := range data.Repos {
        if r.Name == name {
            data.Repos = append(data.Repos[:i], data.Repos[i+1:]...)
            Save()
            return true
        }
    }
    return false
}

func ListRepos() []Repo {
    mu.RLock()
    defer mu.RUnlock()
    return data.Repos
}

const defaultSystemPrompt = `You are a homelab assistant. Based on the user's message and available repositories, determine the appropriate action.

Available repositories and their commands are provided. Match the user's intent to the most relevant repository command.

If the user's request doesn't match any specific repository, respond conversationally as "chat".

Respond with JSON: {"command": "<command_name>", "reasoning": "<brief explanation>"}
`
