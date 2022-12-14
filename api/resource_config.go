package api

import "context"

func (client *Client) GetConfig(ctx context.Context, appName string) (*AppConfig, error) {
	query := `
			query($appName: String!) {
				app(name: $appName) {
					config {
						definition
					}
				}
			}
		`

	req := client.NewRequest(query)
	req.Var("appName", appName)

	data, err := client.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}
	return &data.App.Config, nil
}

func (client *Client) ParseConfig(ctx context.Context, appName string, definition Definition) (*AppConfig, error) {
	query := `
			query($appName: String!, $definition: JSON!) {
				app(name: $appName) {
					parseConfig(definition: $definition) {
						definition
						valid
						errors
						services {
							description
						}
					}
				}
			}
		`

	req := client.NewRequest(query)
	req.Var("appName", appName)
	req.Var("definition", definition)

	data, err := client.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.App.ParseConfig, nil
}

func (client *Client) ValidateConfig(ctx context.Context, appName string, definition Definition) (*AppConfig, error) {
	query := `
			query($definition: JSON!) {
				validateConfig(definition: $definition) {
					definition
					valid
					errors
				}
			}
		`

	req := client.NewRequest(query)
	req.Var("definition", definition)

	data, err := client.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.ValidateConfig, nil
}
