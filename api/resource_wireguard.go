package api

import (
	"context"
	"fmt"
	"os"
)

// GetWireGuardPeerStatus is distinct from the rest of the WireGuard
// accessors because we don't to ask for status routinely, only when
// the user actually needs it (it incurs costs serverside)
func (c *Client) GetWireGuardPeerStatus(ctx context.Context, slug, name string) (*WireGuardPeerStatus, error) {
	req := c.NewRequest(`
query($slug: String!, $name: String!) {
  organization(slug: $slug) {
    wireGuardPeer(name: $name) {
      gatewayStatus
    }
  }
}
`)
	req.Var("slug", slug)
	req.Var("name", name)

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return data.Organization.WireGuardPeer.GatewayStatus, nil
}

func (c *Client) GetWireGuardPeer(ctx context.Context, slug, name string) (*WireGuardPeer, error) {
	req := c.NewRequest(`
query($slug: String!, $name: String!) {
  organization(slug: $slug) {
    wireGuardPeer(name: $name) {
      id
      name
      pubkey
      region
      peerip
    }
  }
}
`)
	req.Var("slug", slug)
	req.Var("name", name)

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	// this graphql code is satanic
	return data.Organization.WireGuardPeer, nil
}

func (c *Client) GetWireGuardPeers(ctx context.Context, slug string) ([]*WireGuardPeer, error) {
	req := c.NewRequest(`
query($slug: String!) {
  organization(slug: $slug) {
    wireGuardPeers {
      nodes {
        id
        name
        pubkey
        region
        peerip
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

	return *data.Organization.WireGuardPeers.Nodes, nil
}

func (c *Client) CreateWireGuardPeer(ctx context.Context, org *Organization, region, name, pubkey string) (*CreatedWireGuardPeer, error) {
	req := c.NewRequest(`
mutation($input: AddWireGuardPeerInput!) {
  addWireGuardPeer(input: $input) {
    peerip
    endpointip
    pubkey
  }
}
`)

	var nats bool

	if os.Getenv("WG_NATS") != "" {
		nats = true
		fmt.Printf("Creating wiregard peer via NATS")
	}

	inputs := map[string]interface{}{
		"organizationId": org.ID,
		"name":           name,
		"pubkey":         pubkey,
		"nats":           nats,
	}

	if region != "" {
		inputs["region"] = region
	}

	req.Var("input", inputs)

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.AddWireGuardPeer, nil
}

func (c *Client) RemoveWireGuardPeer(ctx context.Context, org *Organization, name string) error {
	req := c.NewRequest(`
mutation($input: RemoveWireGuardPeerInput!) {
  removeWireGuardPeer(input: $input) {
    organization {
      id
    }
  }
}
`)
	req.Var("input", map[string]interface{}{
		"organizationId": org.ID,
		"name":           name,
	})

	_, err := c.RunWithContext(ctx, req)

	return err
}

func (c *Client) CreateDelegatedWireGuardToken(ctx context.Context, org *Organization, name string) (*DelegatedWireGuardToken, error) {
	req := c.NewRequest(`
mutation($input: CreateDelegatedWireGuardTokenInput!) {
  createDelegatedWireGuardToken(input: $input) {
    token
  }
}
`)
	req.Var("input", map[string]interface{}{
		"organizationId": org.ID,
		"name":           name,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.CreateDelegatedWireGuardToken, nil
}

func (c *Client) DeleteDelegatedWireGuardToken(ctx context.Context, org *Organization, name, token *string) error {
	query := `
mutation($input: DeleteDelegatedWireGuardTokenInput!) {
  deleteDelegatedWireGuardToken(input: $input) {
    token
  }
}
`

	input := map[string]interface{}{
		"organizationId": org.ID,
	}

	if name != nil {
		input["name"] = *name
	} else {
		input["token"] = *token
	}

	fmt.Printf("%+v\n", input)

	req := c.NewRequest(query)
	req.Var("input", input)

	_, err := c.RunWithContext(ctx, req)

	return err
}

func (c *Client) GetDelegatedWireGuardTokens(ctx context.Context, slug string) ([]*DelegatedWireGuardTokenHandle, error) {
	req := c.NewRequest(`
query($slug: String!) {
  organization(slug: $slug) {
    delegatedWireGuardTokens {
      nodes {
        name
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

	return *data.Organization.DelegatedWireGuardTokens.Nodes, nil
}

func (c *Client) ClosestWireguardGatewayRegion(ctx context.Context) (*Region, error) {
	req := c.NewRequest(`
		query {
			nearestRegion(wireguardGateway: true) {
				code
				name
				gatewayAvailable
			}
		}
`)

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return data.NearestRegion, nil
}

func (c *Client) ValidateWireGuardPeers(ctx context.Context, peerIPs []string) (invalid []string, err error) {
	req := c.NewRequest(`
mutation($input: ValidateWireGuardPeersInput!) {
  validateWireGuardPeers(input: $input) {
		invalidPeerIps
	}
}
`)

	req.Var("input", map[string]interface{}{
		"peerIps": peerIPs,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return data.ValidateWireGuardPeers.InvalidPeerIPs, nil
}
