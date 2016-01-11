package cloudapi

import (
	"net/http"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gocommon/errors"
)

// UpdateMachineMetadata updates the metadata for a given machine.
// Any metadata keys passed in here are created if they do not exist, and
// overwritten if they do.
// See API docs: http://apidocs.joyent.com/cloudapi/#UpdateMachineMetadata
func (c *Client) UpdateMachineMetadata(machineId string, metadata map[string]string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	req := request{
		method:   client.POST,
		url:      makeURL(apiMachines, machineId, apiMetadata),
		reqValue: metadata,
		resp:     &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to update metadata for machine with id %s", machineId)
	}
	return resp, nil
}

// GetMachineMetadata returns the complete set of metadata associated with the
// specified machine.
// See API docs: http://apidocs.joyent.com/cloudapi/#GetMachineMetadata
func (c *Client) GetMachineMetadata(machineId string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	req := request{
		method: client.GET,
		url:    makeURL(apiMachines, machineId, apiMetadata),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get list of metadata for machine with id %s", machineId)
	}
	return resp, nil
}

// DeleteMachineMetadata deletes a single metadata key from the specified machine.
// See API docs: http://apidocs.joyent.com/cloudapi/#DeleteMachineMetadata
func (c *Client) DeleteMachineMetadata(machineId, metadataKey string) error {
	req := request{
		method:         client.DELETE,
		url:            makeURL(apiMachines, machineId, apiMetadata, metadataKey),
		expectedStatus: http.StatusNoContent,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to delete metadata with key %s for machine with id %s", metadataKey, machineId)
	}
	return nil
}

// DeleteAllMachineMetadata deletes all metadata keys from the specified machine.
// See API docs: http://apidocs.joyent.com/cloudapi/#DeleteAllMachineMetadata
func (c *Client) DeleteAllMachineMetadata(machineId string) error {
	req := request{
		method:         client.DELETE,
		url:            makeURL(apiMachines, machineId, apiMetadata),
		expectedStatus: http.StatusNoContent,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to delete metadata for machine with id %s", machineId)
	}
	return nil
}
