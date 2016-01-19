package cloudapi

import (
	"net/http"
	"strconv"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gocommon/errors"
)

type FabricVLAN struct {
	Id          int16  `json:"vlan_id"`               // Number between 0-4095 indicating VLAN Id
	Name        string `json:"name"`                  // Unique name to identify VLAN
	Description string `json:"description,omitempty"` // Optional description of the VLAN
}

type FabricNetwork struct {
	Id               string            `json:"id"`                 // Unique identifier for network
	Name             string            `json:"name"`               // Network name
	Public           bool              `json:"public"`             // Whether or not this is an RFC1918 network
	Fabric           bool              `json:"fabric"`             // Whether this network is on a fabric
	Description      string            `json:"description"`        // Optional description of network
	Subnet           string            `json:"subnet"`             // CIDR formatted string describing network
	ProvisionStartIp string            `json:"provision_start_ip"` // First IP on the network that can be assigned
	ProvisionEndIp   string            `json:"provision_end_ip"`   // Last assignable IP on the network
	Gateway          string            `json:"gateway"`            // Optional Gateway IP
	Resolvers        []string          `json:"resolvers"`          // Array of IP addresses for resolvers
	Routes           map[string]string `json:"routes"`             // Map of CIDR block to Gateway IP Address
	InternetNAT      bool              `json:"internet_nat"`       // If a NAT zone is provisioned at Gateway IP Address
	VLANId           int16             `json:"vlan_id"`            // VLAN network is on
}

type CreateFabricNetworkOpts struct {
	Name             string            `json:"name"`                  // Network name
	Description      string            `json:"description,omitempty"` // Optional description of network
	Subnet           string            `json:"subnet"`                // CIDR formatted string describing network
	ProvisionStartIp string            `json:"provision_start_ip"`    // First IP on the network that can be assigned
	ProvisionEndIp   string            `json:"provision_end_ip"`      // Last assignable IP on the network
	Gateway          string            `json:"gateway,omitempty"`     // Optional Gateway IP
	Resolvers        []string          `json:"resolvers"`             // Array of IP addresses for resolvers
	Routes           map[string]string `json:"routes,omitempty"`      // Map of CIDR block to Gateway IP Address
	InternetNAT      bool              `json:"internet_nat"`          // If a NAT zone is provisioned at Gateway IP Address
}

// See API docs: https://apidocs.joyent.com/cloudapi/#ListFabricVLANs
func (c *Client) ListFabricVLANs() ([]FabricVLAN, error) {
	var resp []FabricVLAN
	req := request{
		method: client.GET,
		url:    apiFabricVLANs,
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get list of fabric VLANs")
	}
	return resp, nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#GetFabricVLAN
func (c *Client) GetFabricVLAN(vlanId int16) (*FabricVLAN, error) {
	var resp FabricVLAN
	req := request{
		method: client.GET,
		url:    makeURL(apiFabricVLANs, strconv.Itoa(int(vlanId))),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get fabric VLAN with id %d", vlanId)
	}
	return &resp, nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#CreateFabricVLAN
func (c *Client) CreateFabricVLAN(vlan FabricVLAN) (*FabricVLAN, error) {
	var resp FabricVLAN
	req := request{
		method:         client.POST,
		url:            apiFabricVLANs,
		reqValue:       vlan,
		resp:           &resp,
		expectedStatus: http.StatusCreated,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to create fabric VLAN: %d - %s", vlan.Id, vlan.Name)
	}
	return &resp, nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#UpdateFabricVLAN
func (c *Client) UpdateFabricVLAN(vlan FabricVLAN) (*FabricVLAN, error) {
	var resp FabricVLAN
	req := request{
		method:         client.PUT,
		url:            makeURL(apiFabricVLANs, strconv.Itoa(int(vlan.Id))),
		reqValue:       vlan,
		resp:           &resp,
		expectedStatus: http.StatusAccepted,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to update fabric VLAN with id %d to %s - %s", vlan.Id, vlan.Name, vlan.Description)
	}
	return &resp, nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#DeleteFabricVLAN
func (c *Client) DeleteFabricVLAN(vlanId int16) error {
	req := request{
		method:         client.DELETE,
		url:            makeURL(apiFabricVLANs, strconv.Itoa(int(vlanId))),
		expectedStatus: http.StatusNoContent,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to delete fabric VLAN with id %d", vlanId)
	}
	return nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#ListFabricNetworks
func (c *Client) ListFabricNetworks(vlanId int16) ([]FabricNetwork, error) {
	var resp []FabricNetwork
	req := request{
		method: client.GET,
		url:    makeURL(apiFabricVLANs, strconv.Itoa(int(vlanId)), apiFabricNetworks),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get list of networks on fabric %d", vlanId)
	}
	return resp, nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#GetFabricNetwork
func (c *Client) GetFabricNetwork(vlanId int16, networkId string) (*FabricNetwork, error) {
	var resp FabricNetwork
	req := request{
		method: client.GET,
		url:    makeURL(apiFabricVLANs, strconv.Itoa(int(vlanId)), apiFabricNetworks, networkId),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get fabric network %s on vlan %d", networkId, vlanId)
	}
	return &resp, nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#CreateFabricNetwork
func (c *Client) CreateFabricNetwork(vlanId int16, opts CreateFabricNetworkOpts) (*FabricNetwork, error) {
	var resp FabricNetwork
	req := request{
		method:         client.POST,
		url:            makeURL(apiFabricVLANs, strconv.Itoa(int(vlanId)), apiFabricNetworks),
		reqValue:       opts,
		resp:           &resp,
		expectedStatus: http.StatusCreated,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to create fabric network %s on vlan %d", opts.Name, vlanId)
	}
	return &resp, nil
}

// See API docs: https://apidocs.joyent.com/cloudapi/#DeleteFabricNetwork
func (c *Client) DeleteFabricNetwork(vlanId int16, networkId string) error {
	req := request{
		method:         client.DELETE,
		url:            makeURL(apiFabricVLANs, strconv.Itoa(int(vlanId)), apiFabricNetworks, networkId),
		expectedStatus: http.StatusNoContent,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to delete fabric network %s on vlan %d", networkId, vlanId)
	}
	return nil
}
