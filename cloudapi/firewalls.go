package cloudapi

import (
	"net/http"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gocommon/errors"
)

// FirewallRule represent a firewall rule that can be specifed for a machine.
type FirewallRule struct {
	Id      string // Unique identifier for the rule
	Enabled bool   // Whether the rule is enabled or not
	Rule    string // Firewall rule in the form 'FROM <target a> TO <target b> <action> <protocol> <port>'
}

// CreateFwRuleOpts represent the option that can be specified
// when creating a new firewall rule.
type CreateFwRuleOpts struct {
	Enabled bool   `json:"enabled"` // Whether to enable the rule or not
	Rule    string `json:"rule"`    // Firewall rule in the form 'FROM <target a> TO <target b> <action> <protocol> <port>'
}

// Lists all the firewall rules on record for a specified account.
// See API docs: http://apidocs.joyent.com/cloudapi/#ListFirewallRules
func (c *Client) ListFirewallRules() ([]FirewallRule, error) {
	var resp []FirewallRule
	req := request{
		method: client.GET,
		url:    apiFirewallRules,
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get list of firewall rules")
	}
	return resp, nil
}

// Returns the specified firewall rule.
// See API docs: http://apidocs.joyent.com/cloudapi/#GetFirewallRule
func (c *Client) GetFirewallRule(fwRuleId string) (*FirewallRule, error) {
	var resp FirewallRule
	req := request{
		method: client.GET,
		url:    makeURL(apiFirewallRules, fwRuleId),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get firewall rule with id %s", fwRuleId)
	}
	return &resp, nil
}

// Creates the firewall rule with the specified options.
// See API docs: http://apidocs.joyent.com/cloudapi/#CreateFirewallRule
func (c *Client) CreateFirewallRule(opts CreateFwRuleOpts) (*FirewallRule, error) {
	var resp FirewallRule
	req := request{
		method:         client.POST,
		url:            apiFirewallRules,
		reqValue:       opts,
		resp:           &resp,
		expectedStatus: http.StatusCreated,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to create firewall rule: %s", opts.Rule)
	}
	return &resp, nil
}

// Updates the specified firewall rule.
// See API docs: http://apidocs.joyent.com/cloudapi/#UpdateFirewallRule
func (c *Client) UpdateFirewallRule(fwRuleId string, opts CreateFwRuleOpts) (*FirewallRule, error) {
	var resp FirewallRule
	req := request{
		method:   client.POST,
		url:      makeURL(apiFirewallRules, fwRuleId),
		reqValue: opts,
		resp:     &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to update firewall rule with id %s to %s", fwRuleId, opts.Rule)
	}
	return &resp, nil
}

// Enables the given firewall rule record if it is disabled.
// See API docs: http://apidocs.joyent.com/cloudapi/#EnableFirewallRule
func (c *Client) EnableFirewallRule(fwRuleId string) (*FirewallRule, error) {
	var resp FirewallRule
	req := request{
		method: client.POST,
		url:    makeURL(apiFirewallRules, fwRuleId, apiFirewallRulesEnable),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to enable firewall rule with id %s", fwRuleId)
	}
	return &resp, nil
}

// Disables the given firewall rule record if it is enabled.
// See API docs: http://apidocs.joyent.com/cloudapi/#DisableFirewallRule
func (c *Client) DisableFirewallRule(fwRuleId string) (*FirewallRule, error) {
	var resp FirewallRule
	req := request{
		method: client.POST,
		url:    makeURL(apiFirewallRules, fwRuleId, apiFirewallRulesDisable),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to disable firewall rule with id %s", fwRuleId)
	}
	return &resp, nil
}

// Removes the given firewall rule record from all the required account machines.
// See API docs: http://apidocs.joyent.com/cloudapi/#DeleteFirewallRule
func (c *Client) DeleteFirewallRule(fwRuleId string) error {
	req := request{
		method:         client.DELETE,
		url:            makeURL(apiFirewallRules, fwRuleId),
		expectedStatus: http.StatusNoContent,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to delete firewall rule with id %s", fwRuleId)
	}
	return nil
}

// Return the list of machines affected by the given firewall rule.
// See API docs: http://apidocs.joyent.com/cloudapi/#ListFirewallRuleMachines
func (c *Client) ListFirewallRuleMachines(fwRuleId string) ([]Machine, error) {
	var resp []Machine
	req := request{
		method: client.GET,
		url:    makeURL(apiFirewallRules, fwRuleId, apiMachines),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get list of machines affected by firewall rule wit id %s", fwRuleId)
	}
	return resp, nil
}
