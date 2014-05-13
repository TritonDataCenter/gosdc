//
// gosdc - Go library to interact with the Joyent CloudAPI
//
//
// Copyright (c) 2013 Joyent Inc.
//
// Written by Daniele Stroppa <daniele.stroppa@joyent.com>
//

package cloudapi_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	gc "launchpad.net/gocheck"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gosdc/cloudapi"
	lc "github.com/joyent/gosdc/localservices/cloudapi"
	"github.com/joyent/gosign/auth"
)

var privateKey []byte

func registerLocalTests(keyName string) {
	var localKeyFile string
	if keyName == "" {
		localKeyFile = os.Getenv("HOME") + "/.ssh/id_rsa"
	} else {
		localKeyFile = keyName
	}
	privateKey, _ = ioutil.ReadFile(localKeyFile)

	gc.Suite(&LocalTests{})
}

type LocalTests struct {
	//LocalTests
	creds      *auth.Credentials
	testClient *cloudapi.Client
	Server     *httptest.Server
	Mux        *http.ServeMux
	oldHandler http.Handler
	cloudapi   *lc.CloudAPI
}

func (s *LocalTests) SetUpSuite(c *gc.C) {
	// Set up the HTTP server.
	s.Server = httptest.NewServer(nil)
	s.oldHandler = s.Server.Config.Handler
	s.Mux = http.NewServeMux()
	s.Server.Config.Handler = s.Mux

	// Set up a Joyent CloudAPI service.
	authentication, err := auth.NewAuth("localtest", string(privateKey), "rsa-sha256")
	c.Assert(err, gc.IsNil)

	s.creds = &auth.Credentials{
		UserAuthentication: authentication,
		SdcKeyId:           "",
		SdcEndpoint:        auth.Endpoint{URL: s.Server.URL},
	}
	s.cloudapi = lc.New(s.creds.SdcEndpoint.URL, s.creds.UserAuthentication.User)
	s.cloudapi.SetupHTTP(s.Mux)
}

func (s *LocalTests) TearDownSuite(c *gc.C) {
	s.Mux = nil
	s.Server.Config.Handler = s.oldHandler
	s.Server.Close()
}

func (s *LocalTests) SetUpTest(c *gc.C) {
	client := client.NewClient(s.creds.SdcEndpoint.URL, cloudapi.DefaultAPIVersion, s.creds, &cloudapi.Logger)
	c.Assert(client, gc.NotNil)
	s.testClient = cloudapi.New(client)
	c.Assert(s.testClient, gc.NotNil)
}

// Helper method to create a test key in the user account
func (s *LocalTests) createKey(c *gc.C) {
	key, err := s.testClient.CreateKey(cloudapi.CreateKeyOpts{Name: "fake-key", Key: testKey})
	c.Assert(err, gc.IsNil)
	c.Assert(key, gc.DeepEquals, &cloudapi.Key{Name: "fake-key", Fingerprint: "", Key: testKey})
}

func (s *LocalTests) deleteKey(c *gc.C) {
	err := s.testClient.DeleteKey("fake-key")
	c.Assert(err, gc.IsNil)
}

// Helper method to create a test virtual machine in the user account
func (s *LocalTests) createMachine(c *gc.C) *cloudapi.Machine {
	machine, err := s.testClient.CreateMachine(cloudapi.CreateMachineOpts{Package: localPackageName, Image: localImageId})
	c.Assert(err, gc.IsNil)
	c.Assert(machine, gc.NotNil)

	// wait for machine to be provisioned
	for !s.pollMachineState(c, machine.Id, "running") {
		time.Sleep(1 * time.Second)
	}

	return machine
}

// Helper method to test the state of a given VM
func (s *LocalTests) pollMachineState(c *gc.C, machineId, state string) bool {
	machineConfig, err := s.testClient.GetMachine(machineId)
	c.Assert(err, gc.IsNil)
	return strings.EqualFold(machineConfig.State, state)
}

// Helper method to delete a test virtual machine once the test has executed
func (s *LocalTests) deleteMachine(c *gc.C, machineId string) {
	err := s.testClient.StopMachine(machineId)
	c.Assert(err, gc.IsNil)

	// wait for machine to be stopped
	for !s.pollMachineState(c, machineId, "stopped") {
		time.Sleep(1 * time.Second)
	}

	err = s.testClient.DeleteMachine(machineId)
	c.Assert(err, gc.IsNil)
}

// Helper method to list virtual machine according to the specified filter
func (s *LocalTests) listMachines(c *gc.C, filter *cloudapi.Filter) {
	var contains bool
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	machines, err := s.testClient.ListMachines(filter)
	c.Assert(err, gc.IsNil)
	c.Assert(machines, gc.NotNil)
	for _, m := range machines {
		if m.Id == testMachine.Id {
			contains = true
			break
		}
	}

	// result
	if !contains {
		c.Fatalf("Obtained machines [%s] do not contain test machine [%s]", machines, *testMachine)
	}
}

// Helper method to create a test firewall rule
func (s *LocalTests) createFirewallRule(c *gc.C) *cloudapi.FirewallRule {
	fwRule, err := s.testClient.CreateFirewallRule(cloudapi.CreateFwRuleOpts{Enabled: false, Rule: testFwRule})
	c.Assert(err, gc.IsNil)
	c.Assert(fwRule, gc.NotNil)
	c.Assert(fwRule.Rule, gc.Equals, testFwRule)
	c.Assert(fwRule.Enabled, gc.Equals, false)
	time.Sleep(10 * time.Second)

	return fwRule
}

// Helper method to a test firewall rule
func (s *LocalTests) deleteFwRule(c *gc.C, fwRuleId string) {
	err := s.testClient.DeleteFirewallRule(fwRuleId)
	c.Assert(err, gc.IsNil)
}

// Keys API
func (s *LocalTests) TestCreateKey(c *gc.C) {
	s.createKey(c)
	s.deleteKey(c)
}

func (s *LocalTests) TestListKeys(c *gc.C) {
	s.createKey(c)
	defer s.deleteKey(c)

	keys, err := s.testClient.ListKeys()
	c.Assert(err, gc.IsNil)
	c.Assert(keys, gc.NotNil)
	fakeKey := cloudapi.Key{Name: "fake-key", Fingerprint: "", Key: testKey}
	for _, k := range keys {
		if c.Check(k, gc.DeepEquals, fakeKey) {
			c.SucceedNow()
		}
	}
	c.Fatalf("Obtained keys [%s] do not contain test key [%s]", keys, fakeKey)
}

func (s *LocalTests) TestGetKeyByName(c *gc.C) {
	s.createKey(c)
	defer s.deleteKey(c)

	key, err := s.testClient.GetKey("fake-key")
	c.Assert(err, gc.IsNil)
	c.Assert(key, gc.NotNil)
	c.Assert(key, gc.DeepEquals, &cloudapi.Key{Name: "fake-key", Fingerprint: "", Key: testKey})
}

/*func (s *LocalTests) TestGetKeyByFingerprint(c *gc.C) {
	s.createKey(c)
	defer s.deleteKey(c)

	key, err := s.testClient.GetKey(testKeyFingerprint)
	c.Assert(err, gc.IsNil)
	c.Assert(key, gc.NotNil)
	c.Assert(key, gc.DeepEquals, &cloudapi.Key{Name: "fake-key", Fingerprint: testKeyFingerprint, Key: testKey})
} */

func (s *LocalTests) TestDeleteKey(c *gc.C) {
	s.createKey(c)

	s.deleteKey(c)
}

// Packages API
func (s *LocalTests) TestListPackages(c *gc.C) {
	pkgs, err := s.testClient.ListPackages(nil)
	c.Assert(err, gc.IsNil)
	c.Assert(pkgs, gc.NotNil)
	for _, pkg := range pkgs {
		c.Check(pkg.Name, gc.FitsTypeOf, string(""))
		c.Check(pkg.Memory, gc.FitsTypeOf, int(0))
		c.Check(pkg.Disk, gc.FitsTypeOf, int(0))
		c.Check(pkg.Swap, gc.FitsTypeOf, int(0))
		c.Check(pkg.VCPUs, gc.FitsTypeOf, int(0))
		c.Check(pkg.Default, gc.FitsTypeOf, bool(false))
		c.Check(pkg.Id, gc.FitsTypeOf, string(""))
		c.Check(pkg.Version, gc.FitsTypeOf, string(""))
		c.Check(pkg.Description, gc.FitsTypeOf, string(""))
		c.Check(pkg.Group, gc.FitsTypeOf, string(""))
	}
}

func (s *LocalTests) TestListPackagesWithFilter(c *gc.C) {
	filter := cloudapi.NewFilter()
	filter.Set("memory", "1024")
	pkgs, err := s.testClient.ListPackages(filter)
	c.Assert(err, gc.IsNil)
	c.Assert(pkgs, gc.NotNil)
	for _, pkg := range pkgs {
		c.Check(pkg.Name, gc.FitsTypeOf, string(""))
		c.Check(pkg.Memory, gc.Equals, 1024)
		c.Check(pkg.Disk, gc.FitsTypeOf, int(0))
		c.Check(pkg.Swap, gc.FitsTypeOf, int(0))
		c.Check(pkg.VCPUs, gc.FitsTypeOf, int(0))
		c.Check(pkg.Default, gc.FitsTypeOf, bool(false))
		c.Check(pkg.Id, gc.FitsTypeOf, string(""))
		c.Check(pkg.Version, gc.FitsTypeOf, string(""))
		c.Check(pkg.Description, gc.FitsTypeOf, string(""))
		c.Check(pkg.Group, gc.FitsTypeOf, string(""))
	}
}

func (s *LocalTests) TestGetPackageFromName(c *gc.C) {
	key, err := s.testClient.GetPackage(localPackageName)
	c.Assert(err, gc.IsNil)
	c.Assert(key, gc.NotNil)
	c.Assert(key, gc.DeepEquals, &cloudapi.Package{
		Name:    "Small",
		Memory:  1024,
		Disk:    16384,
		Swap:    2048,
		VCPUs:   1,
		Default: true,
		Id:      "11223344-1212-abab-3434-aabbccddeeff",
		Version: "1.0.2",
	})
}

func (s *LocalTests) TestGetPackageFromId(c *gc.C) {
	key, err := s.testClient.GetPackage(localPackageId)
	c.Assert(err, gc.IsNil)
	c.Assert(key, gc.NotNil)
	c.Assert(key, gc.DeepEquals, &cloudapi.Package{
		Name:    "Small",
		Memory:  1024,
		Disk:    16384,
		Swap:    2048,
		VCPUs:   1,
		Default: true,
		Id:      "11223344-1212-abab-3434-aabbccddeeff",
		Version: "1.0.2",
	})
}

// Images API
func (s *LocalTests) TestListImages(c *gc.C) {
	imgs, err := s.testClient.ListImages(nil)
	c.Assert(err, gc.IsNil)
	c.Assert(imgs, gc.NotNil)
	for _, img := range imgs {
		c.Check(img.Id, gc.FitsTypeOf, string(""))
		c.Check(img.Name, gc.FitsTypeOf, string(""))
		c.Check(img.OS, gc.FitsTypeOf, string(""))
		c.Check(img.Version, gc.FitsTypeOf, string(""))
		c.Check(img.Type, gc.FitsTypeOf, string(""))
		c.Check(img.Description, gc.FitsTypeOf, string(""))
		c.Check(img.Requirements, gc.FitsTypeOf, map[string]interface{}{"key": "value"})
		c.Check(img.Homepage, gc.FitsTypeOf, string(""))
		c.Check(img.PublishedAt, gc.FitsTypeOf, string(""))
		c.Check(img.Public, gc.FitsTypeOf, string(""))
		c.Check(img.State, gc.FitsTypeOf, string(""))
		c.Check(img.Tags, gc.FitsTypeOf, map[string]string{"key": "value"})
		c.Check(img.EULA, gc.FitsTypeOf, string(""))
		c.Check(img.ACL, gc.FitsTypeOf, []string{"", ""})
	}
}

func (s *LocalTests) TestListImagesWithFilter(c *gc.C) {
	filter := cloudapi.NewFilter()
	filter.Set("os", "smartos")
	imgs, err := s.testClient.ListImages(filter)
	c.Assert(err, gc.IsNil)
	c.Assert(imgs, gc.NotNil)
	for _, img := range imgs {
		c.Check(img.Id, gc.FitsTypeOf, string(""))
		c.Check(img.Name, gc.FitsTypeOf, string(""))
		c.Check(img.OS, gc.Equals, "smartos")
		c.Check(img.Version, gc.FitsTypeOf, string(""))
		c.Check(img.Type, gc.FitsTypeOf, string(""))
		c.Check(img.Description, gc.FitsTypeOf, string(""))
		c.Check(img.Requirements, gc.FitsTypeOf, map[string]interface{}{"key": "value"})
		c.Check(img.Homepage, gc.FitsTypeOf, string(""))
		c.Check(img.PublishedAt, gc.FitsTypeOf, string(""))
		c.Check(img.Public, gc.FitsTypeOf, string(""))
		c.Check(img.State, gc.FitsTypeOf, string(""))
		c.Check(img.Tags, gc.FitsTypeOf, map[string]string{"key": "value"})
		c.Check(img.EULA, gc.FitsTypeOf, string(""))
		c.Check(img.ACL, gc.FitsTypeOf, []string{"", ""})
	}
}

// TODO Add test for deleteImage, exportImage and CreateMachineFormIMage

func (s *LocalTests) TestGetImage(c *gc.C) {
	img, err := s.testClient.GetImage(localImageId)
	c.Assert(err, gc.IsNil)
	c.Assert(img, gc.NotNil)
	c.Assert(img, gc.DeepEquals, &cloudapi.Image{
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
	})
}

// Tests for Machine API
func (s *LocalTests) TestCreateMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	c.Assert(testMachine.Type, gc.Equals, "smartmachine")
	c.Assert(testMachine.Memory, gc.Equals, 1024)
	c.Assert(testMachine.Disk, gc.Equals, 16384)
	c.Assert(testMachine.Package, gc.Equals, localPackageName)
	c.Assert(testMachine.Image, gc.Equals, localImageId)
}

func (s *LocalTests) TestListMachines(c *gc.C) {
	s.listMachines(c, nil)
}

func (s *LocalTests) TestListMachinesWithFilter(c *gc.C) {
	filter := cloudapi.NewFilter()
	filter.Set("memory", "1024")

	s.listMachines(c, filter)
}

/*func (s *LocalTests) TestCountMachines(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	count, err := s.testClient.CountMachines()
	c.Assert(err, gc.IsNil)
	c.Assert(count >= 1, gc.Equals, true)
}*/

func (s *LocalTests) TestGetMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	machine, err := s.testClient.GetMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
	c.Assert(machine, gc.NotNil)
	c.Assert(machine.Equals(*testMachine), gc.Equals, true)
}

func (s *LocalTests) TestStopMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	err := s.testClient.StopMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
}

func (s *LocalTests) TestStartMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	err := s.testClient.StopMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)

	// wait for machine to be stopped
	for !s.pollMachineState(c, testMachine.Id, "stopped") {
		time.Sleep(1 * time.Second)
	}

	err = s.testClient.StartMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
}

func (s *LocalTests) TestRebootMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	err := s.testClient.RebootMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
}

func (s *LocalTests) TestRenameMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	err := s.testClient.RenameMachine(testMachine.Id, "test-machine-renamed")
	c.Assert(err, gc.IsNil)

	renamed, err := s.testClient.GetMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
	c.Assert(renamed.Name, gc.Equals, "test-machine-renamed")
}

func (s *LocalTests) TestResizeMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	err := s.testClient.ResizeMachine(testMachine.Id, "Medium")
	c.Assert(err, gc.IsNil)

	resized, err := s.testClient.GetMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
	c.Assert(resized.Package, gc.Equals, "Medium")
}

func (s *LocalTests) TestListMachinesFirewallRules(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	fwRules, err := s.testClient.ListMachineFirewallRules(testMachine.Id)
	c.Assert(err, gc.IsNil)
	c.Assert(fwRules, gc.NotNil)
}

func (s *LocalTests) TestEnableFirewallMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	err := s.testClient.EnableFirewallMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
}

func (s *LocalTests) TestDisableFirewallMachine(c *gc.C) {
	testMachine := s.createMachine(c)
	defer s.deleteMachine(c, testMachine.Id)

	err := s.testClient.DisableFirewallMachine(testMachine.Id)
	c.Assert(err, gc.IsNil)
}

func (s *LocalTests) TestDeleteMachine(c *gc.C) {
	testMachine := s.createMachine(c)

	s.deleteMachine(c, testMachine.Id)
}

// FirewallRules API
func (s *LocalTests) TestCreateFirewallRule(c *gc.C) {
	testFwRule := s.createFirewallRule(c)

	// cleanup
	s.deleteFwRule(c, testFwRule.Id)
}

func (s *LocalTests) TestListFirewallRules(c *gc.C) {
	testFwRule := s.createFirewallRule(c)
	defer s.deleteFwRule(c, testFwRule.Id)

	rules, err := s.testClient.ListFirewallRules()
	c.Assert(err, gc.IsNil)
	c.Assert(rules, gc.NotNil)
}

func (s *LocalTests) TestGetFirewallRule(c *gc.C) {
	testFwRule := s.createFirewallRule(c)
	defer s.deleteFwRule(c, testFwRule.Id)

	fwRule, err := s.testClient.GetFirewallRule(testFwRule.Id)
	c.Assert(err, gc.IsNil)
	c.Assert(fwRule, gc.NotNil)
	c.Assert((*fwRule), gc.DeepEquals, (*testFwRule))
}

func (s *LocalTests) TestUpdateFirewallRule(c *gc.C) {
	testFwRule := s.createFirewallRule(c)
	defer s.deleteFwRule(c, testFwRule.Id)

	fwRule, err := s.testClient.UpdateFirewallRule(testFwRule.Id, cloudapi.CreateFwRuleOpts{Rule: testUpdatedFwRule})
	c.Assert(err, gc.IsNil)
	c.Assert(fwRule, gc.NotNil)
	c.Assert(fwRule.Rule, gc.Equals, testUpdatedFwRule)
}

func (s *LocalTests) TestEnableFirewallRule(c *gc.C) {
	testFwRule := s.createFirewallRule(c)
	defer s.deleteFwRule(c, testFwRule.Id)

	fwRule, err := s.testClient.EnableFirewallRule((*testFwRule).Id)
	c.Assert(err, gc.IsNil)
	c.Assert(fwRule, gc.NotNil)
}

func (s *LocalTests) TestDisableFirewallRule(c *gc.C) {
	testFwRule := s.createFirewallRule(c)
	defer s.deleteFwRule(c, testFwRule.Id)

	fwRule, err := s.testClient.DisableFirewallRule((*testFwRule).Id)
	c.Assert(err, gc.IsNil)
	c.Assert(fwRule, gc.NotNil)
}

func (s *LocalTests) TestDeleteFirewallRule(c *gc.C) {
	testFwRule := s.createFirewallRule(c)

	s.deleteFwRule(c, testFwRule.Id)
}

// Networks API
func (s *LocalTests) TestListNetworks(c *gc.C) {
	nets, err := s.testClient.ListNetworks()
	c.Assert(err, gc.IsNil)
	c.Assert(nets, gc.NotNil)
}

func (s *LocalTests) TestGetNetwork(c *gc.C) {
	net, err := s.testClient.GetNetwork(localNetworkId)
	c.Assert(err, gc.IsNil)
	c.Assert(net, gc.NotNil)
	c.Assert(net, gc.DeepEquals, &cloudapi.Network{
		Id:          localNetworkId,
		Name:        "Test-Joyent-Public",
		Public:      true,
		Description: "",
	})
}
