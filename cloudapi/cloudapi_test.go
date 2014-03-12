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
	"flag"
	gc "launchpad.net/gocheck"
	"testing"

	"github.com/joyent/gocommon/jpc"
)

const (
	testKey            = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDdArXEuyqVPwJ7uT/QLFYrGLposHGKRP4U1YPuXFFYQMa2Mq9cke6c6YYoHpNU3mVjatHp+sicfQHcO9nPMaWXoIn53kWdldvo0brsqGXXaHcQCjCaSooJiMgG4jDWUmnfySOQA0sEAXcktqmePpLsDlih05mORiueAR1Mglrc6TiVvjd8ZTPhZejMzETVusMweIilE+K7cNjQVxwHId5WVjTRAqRCvZXAIcP2+fzDXTmuKWhSdln19bKz5AEp1jU/eg4D4PuQvwynb9A8Ra2SJnOZ2+9cfDVhrbpzVMty4qQU6WblJNjpLnLpkm8w0isYk2Vr13a+1/N941gFcZaZ daniele@lightman.local"
	testKeyFingerprint = "6b:06:0c:6b:0b:44:67:97:2c:4f:87:28:28:f3:c6:a9"
	packageId          = "d6ca9994-53e7-4adf-a818-aadd3c90a916"
	localPackageId     = "11223344-1212-abab-3434-aabbccddeeff"
	packageName        = "g3-standard-1-smartos"
	localPackageName   = "Small"
	imageId            = "f669428c-a939-11e2-a485-b790efc0f0c1"
	localImageId       = "12345678-a1a1-b2b2-c3c3-098765432100"
	testFwRule         = "FROM subnet 10.35.76.0/24 TO subnet 10.35.101.0/24 ALLOW tcp (PORT 80 AND PORT 443)"
	testUpdatedFwRule  = "FROM subnet 10.35.76.0/24 TO subnet 10.35.101.0/24 ALLOW tcp (port 80 AND port 443 AND port 8080)"
	networkId          = "42325ea0-eb62-44c1-8eb6-0af3e2f83abc"
	localNetworkId     = "123abc4d-0011-aabb-2233-ccdd4455"
)

var live = flag.Bool("live", false, "Include live Joyent Cloud tests")
var keyName = flag.String("key.name", "", "Specify the full path to the private key, defaults to ~/.ssh/id_rsa")

func Test(t *testing.T) {
	if *live {
		creds, err := jpc.CompleteCredentialsFromEnv(*keyName)
		if err != nil {
			t.Fatalf("Error setting up test suite: %s", err.Error())
		}
		registerJoyentCloudTests(creds)
	}
	registerLocalTests(*keyName)
	gc.TestingT(t)
}
