package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/machinebox/graphql"
)

var baseURL string

func SetBaseURL(url string) {
	baseURL = url
}

type Client struct {
	httpClient  *http.Client
	client      *graphql.Client
	accessToken string
}

func NewClient(accessToken string) (*Client, error) {
	if accessToken == "" {
		return nil, errors.New("No api access token available. Please login")
	}

	httpClient, _ := newHTTPClient()

	url := fmt.Sprintf("%s/api/v2/graphql", baseURL)

	client := graphql.NewClient(url, graphql.WithHTTPClient(httpClient))
	return &Client{httpClient, client, accessToken}, nil
}

func (c *Client) NewRequest(q string) *graphql.Request {
	q = compactQueryString(q)
	return graphql.NewRequest(q)
}

func (c *Client) Run(req *graphql.Request) (Query, error) {
	return c.RunWithContext(context.Background(), req)
}

func (c *Client) RunWithContext(ctx context.Context, req *graphql.Request) (Query, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	var resp Query
	err := c.client.Run(ctx, req, &resp)
	return resp, err
}

var compactPattern = regexp.MustCompile(`\s+`)

func compactQueryString(q string) string {
	q = strings.TrimSpace(q)
	return compactPattern.ReplaceAllString(q, " ")
}

func (c *Client) GetCurrentUser() (*User, error) {
	query := `
		query {
			currentUser {
				email
			}
		}
	`

	req := c.NewRequest(query)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.CurrentUser, nil
}

func GetAccessToken(email, password, otp string) (string, error) {
	postData, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]string{
				"email":    email,
				"password": password,
				"otp":      otp,
			},
		},
	})

	url := fmt.Sprintf("%s/api/v1/sessions", baseURL)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(postData))
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 500 {
		return "", errors.New("An unknown server error occured, please try again")
	}

	if resp.StatusCode >= 400 {
		return "", errors.New("Incorrect email and password combination")
	}

	defer resp.Body.Close()

	var result map[string]map[string]map[string]string

	json.NewDecoder(resp.Body).Decode(&result)

	accessToken := result["data"]["attributes"]["access_token"]

	return accessToken, nil
}

type getLogsResponse struct {
	Data []struct {
		Id         string
		Attributes LogEntry
	}
	Meta struct {
		NextToken string `json:"next_token"`
	}
}

func (c *Client) GetAppLogs(appName string, nextToken string, region string, instanceId string) ([]LogEntry, string, error) {

	data := url.Values{}
	data.Set("next_token", nextToken)
	if instanceId != "" {
		data.Set("instance", instanceId)
	}
	if region != "" {
		data.Set("region", region)
	}

	url := fmt.Sprintf("%s/api/v1/apps/%s/logs?%s", baseURL, appName, data.Encode())

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	var result getLogsResponse

	entries := []LogEntry{}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return entries, "", err
	}

	if resp.StatusCode != 200 {
		return entries, "", ErrorFromResp(resp)
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&result)

	for _, d := range result.Data {
		entries = append(entries, d.Attributes)
	}

	return entries, result.Meta.NextToken, nil
}

func (client *Client) GetOrganizations() ([]Organization, error) {
	q := `
		{
			organizations {
				nodes {
					id
					slug
					name
					type
				}
			}
		}
	`

	req := client.NewRequest(q)

	data, err := client.Run(req)
	if err != nil {
		return []Organization{}, err
	}

	return data.Organizations.Nodes, nil
}

func (client *Client) DeployImage(appName, imageTag string) (*Release, error) {
	query := `
			mutation($input: DeployImageInput!) {
				deployImage(input: $input) {
					release {
						id
						version
						reason
						description
						user {
							id
							email
							name
						}
						createdAt
					}
				}
			}
		`

	req := client.NewRequest(query)

	req.Var("input", map[string]string{
		"appId": appName,
		"image": imageTag,
	})

	data, err := client.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.DeployImage.Release, nil
}

func (c *Client) GetApps() ([]App, error) {
	query := `
		query {
			apps(type: "container") {
				nodes {
					id
					name
					organization {
						slug
					}
				}
			}
		}
		`

	req := c.NewRequest(query)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return data.Apps.Nodes, nil
}

func (c *Client) GetApp(appName string) (*App, error) {
	query := `
		query ($appName: String!) {
			app(name: $appName) {
				id
				name
				deployed
				status
				appUrl
				organization {
					slug
				}
				tasks {
					id
					name
					services {
						protocol
						port
						internalPort
						filters
					}
				}
				ipAddresses {
					nodes {
						id
						address
						type
					}
				}
			}
		}
	`

	req := c.NewRequest(query)
	req.Var("appName", appName)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.App, nil
}

func (c *Client) GetAppReleases(appName string, limit int) ([]Release, error) {
	query := `
		query ($appName: String!, $limit: Int!) {
			app(name: $appName) {
				releases(first: $limit) {
					nodes {
						id
						version
						reason
						description
						user {
							id
							email
							name
						}	
						createdAt
					}
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("appName", appName)
	req.Var("limit", limit)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return data.App.Releases.Nodes, nil
}

func (c *Client) SetSecrets(appName string, secrets map[string]string) (*Release, error) {
	query := `
		mutation($input: SetSecretsInput!) {
			setSecrets(input: $input) {
				release {
					id
					version
					reason
					description
					user {
						id
						email
						name
					}
					createdAt
				}
			}
		}
	`

	input := SetSecretsInput{AppID: appName}
	for k, v := range secrets {
		input.Secrets = append(input.Secrets, SetSecretsInputSecret{Key: k, Value: v})
	}

	req := c.NewRequest(query)

	req.Var("input", input)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.SetSecrets.Release, nil
}

func (c *Client) UnsetSecrets(appName string, keys []string) (*Release, error) {
	query := `
		mutation($input: UnsetSecretsInput!) {
			unsetSecrets(input: $input) {
				release {
					id
					version
					reason
					description
					user {
						id
						email
						name
					}
					createdAt
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", UnsetSecretsInput{AppID: appName, Keys: keys})

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.UnsetSecrets.Release, nil
}

func (c *Client) GetAppSecrets(appName string) ([]Secret, error) {
	query := `
		query ($appName: String!) {
			app(name: $appName) {
				secrets {
					name
					digest
					createdAt
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("appName", appName)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return data.App.Secrets, nil
}

func (c *Client) CreateApp(name string, orgId string) (*App, error) {
	query := `
		mutation($input: CreateAppInput!) {
			createApp(input: $input) {
				app {
					id
					name
					organization {
						slug
					}
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", CreateAppInput{
		Name:           name,
		Runtime:        "FIRECRACKER",
		OrganizationID: orgId,
	})

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.CreateApp.App, nil
}

func (c *Client) GetAppWithTasks(appName string) (*App, error) {
	query := `
		query($appName: String!) {
			app(name: $appName) {
				deployed
				tasks {
					id
					name
					status
					servicesSummary
					allocations {
						id
						version
						status
						desiredStatus
						region
						createdAt
					}
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("appName", appName)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.App, nil
}

func (c *Client) CreateSignedUrls(appId string, filename string) (getUrl string, putUrl string, err error) {
	query := `
		mutation($appId: ID!, $filename: String!) {
			createSignedUrl(appId: $appId, filename: $filename) {
				getUrl
				putUrl
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("appId", appId)
	req.Var("filename", filename)

	data, err := c.Run(req)
	if err != nil {
		return "", "", err
	}

	return data.CreateSignedUrl.GetUrl, data.CreateSignedUrl.PutUrl, nil
}

func (c *Client) CreateBuild(appId string, sourceUrl, sourceType string) (*Build, error) {
	query := `
		mutation($appId: ID!, $sourceUrl: String!, $sourceType: UrlSource!) {
			createBuild(appId: $appId, sourceUrl: $sourceUrl, sourceType: $sourceType) {
				build {
					id
					inProgress
					status
					user {
						id
						name
						email
					}
					createdAt
					updatedAt
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("appId", appId)
	req.Var("sourceUrl", sourceUrl)
	req.Var("sourceType", sourceType)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.CreateBuild.Build, nil
}

func (c *Client) ListBuilds(appName string) ([]Build, error) {
	query := `
		query($appName: String!) {
			app(name: $appName) {
				builds {
					nodes {
						id
						inProgress
						status
						user {
							id
							name
							email
						}
						createdAt
						updatedAt
					}
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("appName", appName)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return data.App.Builds.Nodes, nil
}

func (c *Client) GetBuild(buildId string) (*Build, error) {
	query := `
		query($id: ID!) {
			build: node(id: $id) {
				id
				__typename
				... on Build {
					inProgress
					status
					logs
					user {
						id
						name
						email
					}
					createdAt
					updatedAt
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("id", buildId)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return &data.Build, nil
}
