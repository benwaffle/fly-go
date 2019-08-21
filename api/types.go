package api

import "time"

type Query struct {
	Apps struct {
		Nodes []App
	}
	App           App
	CurrentUser   User
	Organizations struct {
		Nodes []Organization
	}

	Build Build

	// mutations
	CreateApp struct {
		App App
	}

	SetSecrets struct {
		Deployment Deployment
		Release    Release
	}

	UnsetSecrets struct {
		Deployment Deployment
		Release    Release
	}

	DeployImage struct {
		Deployment Deployment
		Release    Release
	}

	CreateSignedUrl SignedUrls

	CreateBuild struct {
		Build Build
	}
}

type App struct {
	ID           string
	Name         string
	Status       string
	Version      int
	AppURL       string
	Organization Organization
	Tasks        []Task
	Secrets      []Secret
	Deployments  struct {
		Nodes []Deployment
	}
	Releases struct {
		Nodes []Release
	}
	IPAddresses struct {
		Nodes []IPAddress
	}
	Builds struct {
		Nodes []Build
	}
}

type Organization struct {
	ID   string
	Name string
	Slug string
}

type Task struct {
	ID              string
	Name            string
	Status          string
	ServicesSummary string
	Services        []Service
	Allocations     []Allocation
}

type Service struct {
	ID           string
	Protocol     string
	Port         int
	InternalPort int
	Filters      []string
}

type Allocation struct {
	ID            string
	Version       int
	Status        string
	DesiredStatus string
	Region        string
	CreatedAt     time.Time
}

type IPAddress struct {
	ID      string
	Address string
	Type    string
}

type User struct {
	ID    string
	Name  string
	Email string
}

type Deployment struct {
	ID           string
	Number       int
	CurrentPhase string
	Description  string
	InProgress   bool
	Reason       string
	Status       string
	Trigger      string
	User         User
	CreatedAt    string
	UpdatedAt    string
	Release      struct {
		Version int
	}
}

type Secret struct {
	Name      string
	Digest    string
	CreatedAt time.Time
}

type SetSecretsInput struct {
	AppID   string                  `json:"appId"`
	Secrets []SetSecretsInputSecret `json:"secrets"`
}

type SetSecretsInputSecret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UnsetSecretsInput struct {
	AppID string   `json:"appId"`
	Keys  []string `json:"keys"`
}

type CreateAppInput struct {
	OrganizationID string `json:"organizationId"`
	Runtime        string `json:"runtime"`
	Name           string `json:"name"`
}

type LogEntry struct {
	Timestamp string
	Message   string
	Level     string
	Meta      struct {
		Instance string
		Region   string
	}
}

type Release struct {
	ID          string
	Version     int
	Reason      string
	Description string
	User        User
	CreatedAt   time.Time
}

type Build struct {
	ID         string
	InProgress bool
	Status     string
	User       User
	Logs       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SignedUrls struct {
	GetUrl string
	PutUrl string
}
