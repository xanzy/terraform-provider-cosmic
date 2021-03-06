package cosmic

import (
	"fmt"
	"testing"

	"github.com/MissionCriticalCloud/go-cosmic/v6/cosmic"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccCosmicStaticRoute_basic(t *testing.T) {
	if COSMIC_VPC_OFFERING == "" {
		t.Skip("This test requires an existing VPC offering (set it by exporting COSMIC_VPC_OFFERING)")
	}

	var route cosmic.StaticRoute

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCosmicStaticRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCosmicStaticRoute_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCosmicStaticRouteExists(
						"cosmic_static_route.foo", &route),
					testAccCheckCosmicStaticRouteAttributes(&route),
				),
			},
		},
	})
}

func testAccCheckCosmicStaticRouteExists(n string, route *cosmic.StaticRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Static Route ID is set")
		}

		client := testAccProvider.Meta().(*CosmicClient)
		r, _, err := client.VPC.GetStaticRouteByID(rs.Primary.ID)

		if err != nil {
			return err
		}

		if r.Id != rs.Primary.ID {
			return fmt.Errorf("Static Route not found")
		}

		*route = *r

		return nil
	}
}

func testAccCheckCosmicStaticRouteAttributes(route *cosmic.StaticRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if route.Cidr != "172.16.0.0/16" {
			return fmt.Errorf("Bad CIDR: %s", route.Cidr)
		}

		return nil
	}
}

func testAccCheckCosmicStaticRouteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CosmicClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cosmic_static_route" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No static route ID is set")
		}

		route, _, err := client.VPC.GetStaticRouteByID(rs.Primary.ID)
		if err == nil && route.Id != "" {
			return fmt.Errorf("Static route %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

var testAccCosmicStaticRoute_basic = fmt.Sprintf(`
resource "cosmic_vpc" "foo" {
  name           = "terraform-vpc"
  display_text   = "terraform-vpc"
  cidr           = "10.0.10.0/22"
  network_domain = "terraform-domain"
  vpc_offering   = "%s"
}

resource "cosmic_static_route" "foo" {
  cidr    = "172.16.0.0/16"
  nexthop = "10.0.252.1"
  vpc_id  = "${cosmic_vpc.foo.id}"
}`,
	COSMIC_VPC_OFFERING,
)
