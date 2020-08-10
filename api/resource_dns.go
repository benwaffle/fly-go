package api

func (c *Client) GetDNSZones(slug string) ([]*DNSZone, error) {
	query := `
		query($slug: String!) {
			organization(slug: $slug) {
				dnsZones {
					nodes {
						id
						domain
						createdAt
					}
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("slug", slug)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return *data.Organization.DNSZones.Nodes, nil
}

func (c *Client) FindDNSZone(organizationSlug string, domain string) (*DNSZone, error) {
	query := `
		query($slug: String!, $domain: String!) {
			organization(slug: $slug) {
				dnsZone(domain: $domain) {
					id
					domain
					createdAt
					organization {
						id
						slug
						name
					}
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("slug", organizationSlug)
	req.Var("domain", domain)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	if data.Organization == nil || data.Organization.DNSZone == nil {
		return nil, ErrNotFound
	}

	return data.Organization.DNSZone, nil
}

func (c *Client) CreateDNSZone(organizationID string, domain string) (*DNSZone, error) {
	query := `
		mutation($input: CreateDnsZoneInput!) {
			createDnsZone(input: $input) {
				zone {
					id
					domain
					createdAt
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]interface{}{
		"organizationId": organizationID,
		"domain":         domain,
	})

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return data.CreateDnsZone.Zone, nil
}

func (c *Client) DeleteDNSZone(zoneID string) error {
	query := `
		mutation($input: DeleteDnsZoneInput!) {
			deleteDnsZone(input: $input) {
				clientMutationId
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]interface{}{
		"dnsZoneId": zoneID,
	})

	_, err := c.Run(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetDNSRecords(zoneID string) ([]*DNSRecord, error) {
	query := `
		query($zoneId: ID!) {
			dnsZone: node(id: $zoneId) {
				... on DnsZone {
					records {
						nodes {
							id
							fqdn
							name
							type
							ttl
							values
							isApex
							isWildcard
							isSystem
							createdAt
							updatedAt
						}
					}
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("zoneId", zoneID)

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	if data.DNSZone == nil {
		return nil, ErrNotFound
	}

	return *data.DNSZone.Records.Nodes, nil
}

func (c *Client) ExportDNSRecords(zoneID string) (string, error) {
	query := `
		mutation($input: ExportDnsZoneInput!) {
			exportDnsZone(input: $input) {
				contents
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]interface{}{
		"dnsZoneId": zoneID,
	})

	data, err := c.Run(req)
	if err != nil {
		return "", err
	}

	return data.ExportDnsZone.Contents, nil
}

func (c *Client) ImportDNSRecords(zoneID string, zonefile string) ([]ImportDnsRecordTypeResult, error) {
	query := `
		mutation($input: ImportDnsZoneInput!) {
			importDnsZone(input: $input) {
				results {
					created
					deleted
					updated
					skipped
					type
				}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]interface{}{
		"dnsZoneId": zoneID,
		"zonefile":  zonefile,
	})

	data, err := c.Run(req)
	if err != nil {
		return nil, err
	}

	return data.ImportDnsZone.Results, nil
}
