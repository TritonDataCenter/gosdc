//
// gosdc - Go library to interact with the Joyent CloudAPI
//
// CloudAPI double testing service - internal direct API implementation
//
// Copyright (c) Joyent Inc.
//

package cloudapi

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/joyent/gosdc/cloudapi"
	"github.com/joyent/gosdc/localservices"
)

var (
	separator       = "/"
	packagesFilters = []string{"name", "memory", "disk", "swap", "version", "vcpus", "group"}
	imagesFilters   = []string{"name", "os", "version", "public", "state", "owner", "type"}
	machinesFilters = []string{"type", "name", "image", "state", "memory", "tombstone", "limit", "offset", "credentials"}
)

// CloudAPI is the API test double
type CloudAPI struct {
	localservices.ServiceInstance
	keys          []cloudapi.Key
	packages      []cloudapi.Package
	images        []cloudapi.Image
	machines      []*cloudapi.Machine
	machineFw     map[string]bool
	snapshots     map[string][]cloudapi.Snapshot
	firewallRules []*cloudapi.FirewallRule
	networks      []cloudapi.Network
}

// New makes a new *CloudAPI service with the given information
func New(serviceURL, userAccount string) *CloudAPI {
	URL, err := url.Parse(serviceURL)
	if err != nil {
		panic(err)
	}
	hostname := URL.Host
	if !strings.HasSuffix(hostname, separator) {
		hostname += separator
	}

	var (
		keys          []cloudapi.Key
		machines      []*cloudapi.Machine
		machineFw     = map[string]bool{}
		snapshots     = map[string][]cloudapi.Snapshot{}
		firewallRules []*cloudapi.FirewallRule
	)

	cloudapiService := &CloudAPI{
		keys:          keys,
		packages:      initPackages(),
		images:        initImages(),
		machines:      machines,
		machineFw:     machineFw,
		snapshots:     snapshots,
		firewallRules: firewallRules,
		networks: []cloudapi.Network{
			{Id: "123abc4d-0011-aabb-2233-ccdd4455", Name: "Test-Joyent-Public", Public: true},
			{Id: "456def0a-33ff-7f8e-9a0b-33bb44cc", Name: "Test-Joyent-Private", Public: false},
		},
		ServiceInstance: localservices.ServiceInstance{
			Scheme:      URL.Scheme,
			Hostname:    hostname,
			UserAccount: userAccount,
		},
	}

	return cloudapiService
}

func initPackages() []cloudapi.Package {
	return []cloudapi.Package{
		{
			Name:    "Micro",
			Memory:  512,
			Disk:    8192,
			Swap:    1024,
			VCPUs:   1,
			Default: false,
			Id:      "12345678-aaaa-bbbb-cccc-000000000000",
			Version: "1.0.0",
		},
		{
			Name:    "Small",
			Memory:  1024,
			Disk:    16384,
			Swap:    2048,
			VCPUs:   1,
			Default: true,
			Id:      "11223344-1212-abab-3434-aabbccddeeff",
			Version: "1.0.2",
		},
		{
			Name:    "Medium",
			Memory:  2048,
			Disk:    32768,
			Swap:    4096,
			VCPUs:   2,
			Default: false,
			Id:      "aabbccdd-abcd-abcd-abcd-112233445566",
			Version: "1.0.4",
		},
		{
			Name:    "Large",
			Memory:  4096,
			Disk:    65536,
			Swap:    16384,
			VCPUs:   4,
			Default: false,
			Id:      "00998877-dddd-eeee-ffff-111111111111",
			Version: "1.0.1",
		},
	}
}

func initImages() []cloudapi.Image {
	return []cloudapi.Image{
		{
			Id:          "12345678-a1a1-b2b2-c3c3-098765432100",
			Name:        "SmartOS Std",
			OS:          "smartos",
			Version:     "13.3.1",
			Type:        "smartmachine",
			Description: "Test SmartOS image (32 bit)",
			Homepage:    "http://test.joyent.com/Standard_Instance",
			PublishedAt: "2014-01-08T17:42:31Z",
			Public:      "true",
			State:       "active",
		},
		{
			Id:          "12345678-b1b1-a4a4-d8d8-111111999999",
			Name:        "standard32",
			OS:          "smartos",
			Version:     "13.3.1",
			Type:        "smartmachine",
			Description: "Test SmartOS image (64 bit)",
			Homepage:    "http://test.joyent.com/Standard_Instance",
			PublishedAt: "2014-01-08T17:43:16Z",
			Public:      "true",
			State:       "active",
		},
		{
			Id:          "a1b2c3d4-0011-2233-4455-0f1e2d3c4b5a",
			Name:        "centos6.4",
			OS:          "linux",
			Version:     "2.4.1",
			Type:        "virtualmachine",
			Description: "Test CentOS 6.4 image (64 bit)",
			PublishedAt: "2014-01-02T10:58:31Z",
			Public:      "true",
			State:       "active",
		},
		{
			Id:          "11223344-0a0a-ff99-11bb-0a1b2c3d4e5f",
			Name:        "ubuntu12.04",
			OS:          "linux",
			Version:     "2.3.1",
			Type:        "virtualmachine",
			Description: "Test Ubuntu 12.04 image (64 bit)",
			PublishedAt: "2014-01-20T16:12:31Z",
			Public:      "true",
			State:       "active",
		},
		{
			Id:          "11223344-0a0a-ee88-22ab-00aa11bb22cc",
			Name:        "ubuntu12.10",
			OS:          "linux",
			Version:     "2.3.2",
			Type:        "virtualmachine",
			Description: "Test Ubuntu 12.10 image (64 bit)",
			PublishedAt: "2014-01-20T16:12:31Z",
			Public:      "true",
			State:       "active",
		},
		{
			Id:          "11223344-0a0a-dd77-33cd-abcd1234e5f6",
			Name:        "ubuntu13.04",
			OS:          "linux",
			Version:     "2.2.8",
			Type:        "virtualmachine",
			Description: "Test Ubuntu 13.04 image (64 bit)",
			PublishedAt: "2014-01-20T16:12:31Z",
			Public:      "true",
			State:       "active",
		},
	}
}

func generatePublicIPAddress() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("32.151.%d.%d", r.Intn(255), r.Intn(255))
}

func generatePrivateIPAddress() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("10.201.%d.%d", r.Intn(255), r.Intn(255))
}

func contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

// Keys APIs

// ListKeys lists keys in the double
func (c *CloudAPI) ListKeys() ([]cloudapi.Key, error) {
	if err := c.ProcessFunctionHook(c); err != nil {
		return nil, err
	}

	return c.keys, nil
}

// GetKey gets a single key from the double by name
func (c *CloudAPI) GetKey(keyName string) (*cloudapi.Key, error) {
	if err := c.ProcessFunctionHook(c, keyName); err != nil {
		return nil, err
	}

	for _, key := range c.keys {
		if key.Name == keyName {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("Key %s not found", keyName)
}

// CreateKey creates a new key in the double
func (c *CloudAPI) CreateKey(keyName, key string) (*cloudapi.Key, error) {
	if err := c.ProcessFunctionHook(c, keyName, key); err != nil {
		return nil, err
	}

	// check if key already exists or keyName already in use
	for _, k := range c.keys {
		if k.Name == keyName {
			return nil, fmt.Errorf("Key name %s already in use", keyName)
		}
		if k.Key == key {
			return nil, fmt.Errorf("Key %s already exists", key)
		}
	}

	newKey := cloudapi.Key{Name: keyName, Fingerprint: "", Key: key}
	c.keys = append(c.keys, newKey)

	return &newKey, nil
}

// DeleteKey deletes an existing key from the double
func (c *CloudAPI) DeleteKey(keyName string) error {
	if err := c.ProcessFunctionHook(c, keyName); err != nil {
		return err
	}

	for i, key := range c.keys {
		if key.Name == keyName {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("Key %s not found", keyName)
}

// Packages APIs

// ListPackages lists packages in the double
func (c *CloudAPI) ListPackages(filters map[string]string) ([]cloudapi.Package, error) {
	if err := c.ProcessFunctionHook(c, filters); err != nil {
		return nil, err
	}

	availablePackages := c.packages

	if filters != nil {
		for k, f := range filters {
			// check if valid filter
			if contains(packagesFilters, k) {
				pkgs := []cloudapi.Package{}
				// filter from availablePackages and add to pkgs
				for _, p := range availablePackages {
					if k == "name" && p.Name == f {
						pkgs = append(pkgs, p)
					} else if k == "memory" {
						i, err := strconv.Atoi(f)
						if err == nil && p.Memory == i {
							pkgs = append(pkgs, p)
						}
					} else if k == "disk" {
						i, err := strconv.Atoi(f)
						if err == nil && p.Disk == i {
							pkgs = append(pkgs, p)
						}
					} else if k == "swap" {
						i, err := strconv.Atoi(f)
						if err == nil && p.Swap == i {
							pkgs = append(pkgs, p)
						}
					} else if k == "version" && p.Version == f {
						pkgs = append(pkgs, p)
					} else if k == "vcpus" {
						i, err := strconv.Atoi(f)
						if err == nil && p.VCPUs == i {
							pkgs = append(pkgs, p)
						}
					} else if k == "group" && p.Group == f {
						pkgs = append(pkgs, p)
					}
				}
				availablePackages = pkgs
			}
		}
	}

	return availablePackages, nil
}

// GetPackage gets a single package in the double
func (c *CloudAPI) GetPackage(packageName string) (*cloudapi.Package, error) {
	if err := c.ProcessFunctionHook(c, packageName); err != nil {
		return nil, err
	}

	for _, pkg := range c.packages {
		if pkg.Name == packageName {
			return &pkg, nil
		}
		if pkg.Id == packageName {
			return &pkg, nil
		}
	}

	return nil, fmt.Errorf("Package %s not found", packageName)
}

// Images APIs

// ListImages returns a list of images in the double
func (c *CloudAPI) ListImages(filters map[string]string) ([]cloudapi.Image, error) {
	if err := c.ProcessFunctionHook(c, filters); err != nil {
		return nil, err
	}

	availableImages := c.images

	if filters != nil {
		for k, f := range filters {
			// check if valid filter
			if contains(imagesFilters, k) {
				imgs := []cloudapi.Image{}
				// filter from availableImages and add to imgs
				for _, i := range availableImages {
					if k == "name" && i.Name == f {
						imgs = append(imgs, i)
					} else if k == "os" && i.OS == f {
						imgs = append(imgs, i)
					} else if k == "version" && i.Version == f {
						imgs = append(imgs, i)
					} else if k == "public" && i.Public == f {
						imgs = append(imgs, i)
					} else if k == "state" && i.State == f {
						imgs = append(imgs, i)
					} else if k == "owner" && i.Owner == f {
						imgs = append(imgs, i)
					} else if k == "type" && i.Type == f {
						imgs = append(imgs, i)
					}
				}
				availableImages = imgs
			}
		}
	}

	return availableImages, nil
}

// GetImage gets a single image by name from the double
func (c *CloudAPI) GetImage(imageID string) (*cloudapi.Image, error) {
	if err := c.ProcessFunctionHook(c, imageID); err != nil {
		return nil, err
	}

	for _, image := range c.images {
		if image.Id == imageID {
			return &image, nil
		}
	}

	return nil, fmt.Errorf("Image %s not found", imageID)
}

// Machine APIs

// ListMachines returns a list of machines in the double
func (c *CloudAPI) ListMachines(filters map[string]string) ([]*cloudapi.Machine, error) {
	if err := c.ProcessFunctionHook(c, filters); err != nil {
		return nil, err
	}

	availableMachines := c.machines

	if filters != nil {
		for k, f := range filters {
			// check if valid filter
			if contains(machinesFilters, k) {
				machines := []*cloudapi.Machine{}
				// filter from availableMachines and add to machines
				for _, m := range availableMachines {
					if k == "name" && m.Name == f {
						machines = append(machines, m)
					} else if k == "type" && m.Type == f {
						machines = append(machines, m)
					} else if k == "state" && m.State == f {
						machines = append(machines, m)
					} else if k == "image" && m.Image == f {
						machines = append(machines, m)
					} else if k == "memory" {
						i, err := strconv.Atoi(f)
						if err == nil && m.Memory == i {
							machines = append(machines, m)
						}
					} else if strings.HasPrefix(k, "tags.") {
						for t, v := range m.Tags {
							if t == k[strings.Index(k, ".")+1:] && v == f {
								machines = append(machines, m)
							}
						}
					}
				}
				availableMachines = machines
			}
		}
	}

	return availableMachines, nil
}

// CountMachines returns a count of machines the double knows about
func (c *CloudAPI) CountMachines() (int, error) {
	if err := c.ProcessFunctionHook(c); err != nil {
		return 0, err
	}

	return len(c.machines), nil
}

// GetMachine gets a single machine by ID from the double
func (c *CloudAPI) GetMachine(machineID string) (*cloudapi.Machine, error) {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return nil, err
	}

	for _, machine := range c.machines {
		if machine.Id == machineID {
			return machine, nil
		}
	}

	return nil, fmt.Errorf("Machine %s not found", machineID)
}

// CreateMachine creates a new machine in the double. It will be running immediately.
func (c *CloudAPI) CreateMachine(name, pkg, image string, networks []string, metadata, tags map[string]string) (*cloudapi.Machine, error) {
	if err := c.ProcessFunctionHook(c, name, pkg, image); err != nil {
		return nil, err
	}

	machineID, err := localservices.NewUUID()
	if err != nil {
		return nil, err
	}

	mPkg, err := c.GetPackage(pkg)
	if err != nil {
		return nil, err
	}

	mImg, err := c.GetImage(image)
	if err != nil {
		return nil, err
	}

	mNetworks := []string{}
	for _, network := range networks {
		mNetwork, err := c.GetNetwork(network)
		if err != nil {
			return nil, err
		}

		mNetworks = append(mNetworks, mNetwork.Id)
	}

	publicIP := generatePublicIPAddress()

	newMachine := &cloudapi.Machine{
		Id:        machineID,
		Name:      name,
		Type:      mImg.Type,
		State:     "running",
		Memory:    mPkg.Memory,
		Disk:      mPkg.Disk,
		IPs:       []string{publicIP, generatePrivateIPAddress()},
		Created:   time.Now().Format("2013-11-26T19:47:13.448Z"),
		Package:   pkg,
		Image:     image,
		Metadata:  metadata,
		Tags:      tags,
		PrimaryIP: publicIP,
		Networks:  mNetworks,
	}
	c.machines = append(c.machines, newMachine)

	return newMachine, nil
}

// StopMachine changes a machine's status to "stopped"
func (c *CloudAPI) StopMachine(machineID string) error {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return err
	}

	for _, machine := range c.machines {
		if machine.Id == machineID {
			machine.State = "stopped"
			machine.Updated = time.Now().Format("2013-11-26T19:47:13.448Z")
			return nil
		}
	}

	return fmt.Errorf("Machine %s not found", machineID)
}

// StartMachine changes a machine's state to "running"
func (c *CloudAPI) StartMachine(machineID string) error {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return err
	}

	for _, machine := range c.machines {
		if machine.Id == machineID {
			machine.State = "running"
			machine.Updated = time.Now().Format("2013-11-26T19:47:13.448Z")
			return nil
		}
	}

	return fmt.Errorf("Machine %s not found", machineID)
}

// RebootMachine changes a machine's state to "running" and updates Updated
func (c *CloudAPI) RebootMachine(machineID string) error {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return err
	}

	for _, machine := range c.machines {
		if machine.Id == machineID {
			machine.State = "running"
			machine.Updated = time.Now().Format("2013-11-26T19:47:13.448Z")
			return nil
		}
	}

	return fmt.Errorf("Machine %s not found", machineID)
}

// ResizeMachine changes a machine's package to a new size. Unlike the real API,
// this method lets you downsize machines.
func (c *CloudAPI) ResizeMachine(machineID, packageName string) error {
	if err := c.ProcessFunctionHook(c, machineID, packageName); err != nil {
		return err
	}

	mPkg, err := c.GetPackage(packageName)
	if err != nil {
		return err
	}

	for _, machine := range c.machines {
		if machine.Id == machineID {
			machine.Package = packageName
			machine.Memory = mPkg.Memory
			machine.Disk = mPkg.Disk
			machine.Updated = time.Now().Format("2013-11-26T19:47:13.448Z")
			return nil
		}
	}

	return fmt.Errorf("Machine %s not found", machineID)
}

// RenameMachine changes a machine's name
func (c *CloudAPI) RenameMachine(machineID, newName string) error {
	if err := c.ProcessFunctionHook(c, machineID, newName); err != nil {
		return err
	}

	for _, machine := range c.machines {
		if machine.Id == machineID {
			machine.Name = newName
			machine.Updated = time.Now().Format("2013-11-26T19:47:13.448Z")
			return nil
		}
	}

	return fmt.Errorf("Machine %s not found", machineID)
}

// ListMachineFirewallRules returns a list of firewall rules that apply to the
// given machine
func (c *CloudAPI) ListMachineFirewallRules(machineID string) ([]*cloudapi.FirewallRule, error) {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return nil, err
	}

	fwRules := []*cloudapi.FirewallRule{}
	for _, r := range c.firewallRules {
		vm := "vm " + machineID
		if strings.Contains(r.Rule, vm) {
			fwRules = append(fwRules, r)
		}
	}

	return fwRules, nil
}

// EnableFirewallMachine enables the firewall for the given machine
func (c *CloudAPI) EnableFirewallMachine(machineID string) error {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return err
	}

	c.machineFw[machineID] = true

	return nil
}

// DisableFirewallMachine disables the firewall for the given machine
func (c *CloudAPI) DisableFirewallMachine(machineID string) error {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return err
	}

	c.machineFw[machineID] = false

	return nil
}

// DeleteMachine deletes the given machine from the double
func (c *CloudAPI) DeleteMachine(machineID string) error {
	if err := c.ProcessFunctionHook(c, machineID); err != nil {
		return err
	}

	for i, machine := range c.machines {
		if machine.Id == machineID {
			if machine.State == "stopped" {
				c.machines = append(c.machines[:i], c.machines[i+1:]...)
				return nil
			}

			return fmt.Errorf("Cannot Delete machine %s, machine is not stopped.", machineID)
		}
	}

	return fmt.Errorf("Machine %s not found", machineID)
}

// FirewallRule APIs

// ListFirewallRules gets a list of firewall rules from the double
func (c *CloudAPI) ListFirewallRules() ([]*cloudapi.FirewallRule, error) {
	if err := c.ProcessFunctionHook(c); err != nil {
		return nil, err
	}

	return c.firewallRules, nil
}

// GetFirewallRule gets a single firewall rule by ID
func (c *CloudAPI) GetFirewallRule(fwRuleID string) (*cloudapi.FirewallRule, error) {
	if err := c.ProcessFunctionHook(c, fwRuleID); err != nil {
		return nil, err
	}

	for _, r := range c.firewallRules {
		if strings.EqualFold(r.Id, fwRuleID) {
			return r, nil
		}
	}

	return nil, fmt.Errorf("Firewall rule %s not found", fwRuleID)
}

// CreateFirewallRule creates a new firewall rule and returns it
func (c *CloudAPI) CreateFirewallRule(rule string, enabled bool) (*cloudapi.FirewallRule, error) {
	if err := c.ProcessFunctionHook(c, rule, enabled); err != nil {
		return nil, err
	}

	fwRuleID, err := localservices.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("Error creating firewall rule: %q", err)
	}

	fwRule := &cloudapi.FirewallRule{Id: fwRuleID, Rule: rule, Enabled: enabled}
	c.firewallRules = append(c.firewallRules, fwRule)

	return fwRule, nil
}

// UpdateFirewallRule makes changes to a given firewall rule
func (c *CloudAPI) UpdateFirewallRule(fwRuleID, rule string, enabled bool) (*cloudapi.FirewallRule, error) {
	if err := c.ProcessFunctionHook(c, fwRuleID, rule, enabled); err != nil {
		return nil, err
	}

	for _, r := range c.firewallRules {
		if strings.EqualFold(r.Id, fwRuleID) {
			r.Rule = rule
			r.Enabled = enabled
			return r, nil
		}
	}

	return nil, fmt.Errorf("Firewall rule %s not found", fwRuleID)
}

// EnableFirewallRule enables the given firewall rule
func (c *CloudAPI) EnableFirewallRule(fwRuleID string) (*cloudapi.FirewallRule, error) {
	if err := c.ProcessFunctionHook(c, fwRuleID); err != nil {
		return nil, err
	}

	for _, r := range c.firewallRules {
		if strings.EqualFold(r.Id, fwRuleID) {
			r.Enabled = true
			return r, nil
		}
	}

	return nil, fmt.Errorf("Firewall rule %s not found", fwRuleID)
}

// DisableFirewallRule disables the given firewall rule
func (c *CloudAPI) DisableFirewallRule(fwRuleID string) (*cloudapi.FirewallRule, error) {
	if err := c.ProcessFunctionHook(c, fwRuleID); err != nil {
		return nil, err
	}

	for _, r := range c.firewallRules {
		if strings.EqualFold(r.Id, fwRuleID) {
			r.Enabled = false
			return r, nil
		}
	}

	return nil, fmt.Errorf("Firewall rule %s not found", fwRuleID)
}

// DeleteFirewallRule deletes the given firewall rule
func (c *CloudAPI) DeleteFirewallRule(fwRuleID string) error {
	if err := c.ProcessFunctionHook(c, fwRuleID); err != nil {
		return err
	}

	for i, r := range c.firewallRules {
		if strings.EqualFold(r.Id, fwRuleID) {
			c.firewallRules = append(c.firewallRules[:i], c.firewallRules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("Firewall rule %s not found", fwRuleID)
}

// ListFirewallRuleMachines should list the machines that are affected by a
// given firewall rule. In this double, it just returns all the machines.
func (c *CloudAPI) ListFirewallRuleMachines(fwRuleID string) ([]*cloudapi.Machine, error) {
	if err := c.ProcessFunctionHook(c, fwRuleID); err != nil {
		return nil, err
	}

	return c.machines, nil
}

// Networks API

// ListNetworks returns a list of networks that the double knows about
func (c *CloudAPI) ListNetworks() ([]cloudapi.Network, error) {
	if err := c.ProcessFunctionHook(c); err != nil {
		return nil, err
	}

	return c.networks, nil
}

// GetNetwork gets a network by ID
func (c *CloudAPI) GetNetwork(networkID string) (*cloudapi.Network, error) {
	if err := c.ProcessFunctionHook(c, networkID); err != nil {
		return nil, err
	}

	for _, n := range c.networks {
		if strings.EqualFold(n.Id, networkID) {
			return &n, nil
		}
	}

	return nil, fmt.Errorf("Network %s not found", networkID)
}
