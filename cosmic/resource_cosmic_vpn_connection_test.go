package cosmic

import (
	"fmt"
	"testing"

	"github.com/MissionCriticalCloud/go-cosmic/v6/cosmic"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccCosmicVPNConnection_basic(t *testing.T) {
	if COSMIC_VPC_OFFERING == "" {
		t.Skip("This test requires an existing VPC offering (set it by exporting COSMIC_VPC_OFFERING)")
	}

	var vpnConnection cosmic.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCosmicVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCosmicVPNConnection_basic(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCosmicVPNConnectionExists(
						"cosmic_vpn_connection.foo-bar", &vpnConnection),
					testAccCheckCosmicVPNConnectionExists(
						"cosmic_vpn_connection.bar-foo", &vpnConnection),
				),
			},
		},
	})
}

func testAccCheckCosmicVPNConnectionExists(n string, vpnConnection *cosmic.VpnConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPN Connection ID is set")
		}

		client := testAccProvider.Meta().(*CosmicClient)
		v, _, err := client.VPN.GetVpnConnectionByID(rs.Primary.ID)

		if err != nil {
			return err
		}

		if v.Id != rs.Primary.ID {
			return fmt.Errorf("VPN Connection not found")
		}

		*vpnConnection = *v

		return nil
	}
}

func testAccCheckCosmicVPNConnectionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CosmicClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cosmic_vpn_connection" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPN Connection ID is set")
		}

		_, _, err := client.VPN.GetVpnConnectionByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("VPN Connection %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCosmicVPNConnection_basic(rand string) string {
	return fmt.Sprintf(`
resource "cosmic_vpc" "foo" {
  name         = "terraform-vpc-foo"
  cidr         = "10.0.10.0/22"
  vpc_offering = "%s"
}

resource "cosmic_vpc" "bar" {
  name         = "terraform-vpc-bar"
  cidr         = "10.0.20.0/22"
  vpc_offering = "%s"
}

resource "cosmic_vpn_gateway" "foo" {
  vpc_id = "${cosmic_vpc.foo.id}"
}

resource "cosmic_vpn_gateway" "bar" {
  vpc_id = "${cosmic_vpc.bar.id}"
}

resource "cosmic_vpn_customer_gateway" "foo" {
  name       = "terraform-foo-%s"
  cidr_list  = ["${cosmic_vpc.foo.cidr}"]
  esp_policy = "aes256-sha1"
  gateway    = "${cosmic_vpn_gateway.foo.public_ip}"
  ike_policy = "aes256-sha1;modp1024"
  ipsec_psk  = "terraform"
}

resource "cosmic_vpn_customer_gateway" "bar" {
  name       = "terraform-bar-%s"
  cidr_list  = ["${cosmic_vpc.bar.cidr}"]
  esp_policy = "aes256-sha1"
  gateway    = "${cosmic_vpn_gateway.bar.public_ip}"
  ike_policy = "aes256-sha1;modp1024"
  ipsec_psk  = "terraform"
}

resource "cosmic_vpn_connection" "foo-bar" {
  customer_gateway_id = "${cosmic_vpn_customer_gateway.foo.id}"
  vpn_gateway_id      = "${cosmic_vpn_gateway.bar.id}"
}

resource "cosmic_vpn_connection" "bar-foo" {
  customer_gateway_id = "${cosmic_vpn_customer_gateway.bar.id}"
  vpn_gateway_id      = "${cosmic_vpn_gateway.foo.id}"
}
`,
		COSMIC_VPC_OFFERING,
		COSMIC_VPC_OFFERING,
		rand,
		rand,
	)
}
