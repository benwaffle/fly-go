package api

import "context"

func (c *Client) GetLoggedCertificates(ctx context.Context, slug string) ([]LoggedCertificate, error) {
	req := c.NewRequest(`
query($slug: String!) { 
  organization(slug: $slug) { 
    loggedCertificates { 
      nodes { 
        root
        cert
      } 
    }
  } 
}
`)
	req.Var("slug", slug)

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return data.Organization.LoggedCertificates.Nodes, nil
}

func (c *Client) EstablishSSHKey(ctx context.Context, org *Organization, override bool) (*SSHCertificate, error) {
	req := c.NewRequest(`
mutation($input: EstablishSSHKeyInput!) { 
  establishSshKey(input: $input) { 
    certificate
  } 
}
`)
	req.Var("input", map[string]interface{}{
		"organizationId": org.ID,
		"override":       override,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.EstablishSSHKey, nil
}

func (c *Client) IssueSSHCertificate(ctx context.Context, org *Organization, email string, username *string, valid_hours *int) (*IssuedCertificate, error) {
	req := c.NewRequest(`
mutation($input: IssueCertificateInput!) { 
  issueCertificate(input: $input) { 
    certificate, key
  } 
}
`)
	inputs := map[string]interface{}{
		"organizationId": org.ID,
		"email":          email,
	}

	if username != nil {
		inputs["username"] = *username
	}

	if valid_hours != nil {
		inputs["validHours"] = *valid_hours
	}

	req.Var("input", inputs)

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.IssueCertificate, nil
}
