package cosmic

import (
	"fmt"
	"log"
	"strings"

	"github.com/MissionCriticalCloud/go-cosmic/v6/cosmic"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCosmicVPNGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceCosmicVPNGatewayCreate,
		Read:   resourceCosmicVPNGatewayRead,
		Delete: resourceCosmicVPNGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"public_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCosmicVPNGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cosmic.CosmicClient)

	vpcid := d.Get("vpc_id").(string)
	p := cs.VPN.NewCreateVpnGatewayParams(vpcid)

	// Create the new VPN Gateway
	v, err := cs.VPN.CreateVpnGateway(p)
	if err != nil {
		return fmt.Errorf("Error creating VPN Gateway for VPC ID %s: %s", vpcid, err)
	}

	d.SetId(v.Id)

	return resourceCosmicVPNGatewayRead(d, meta)
}

func resourceCosmicVPNGatewayRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cosmic.CosmicClient)

	// Get the VPN Gateway details
	v, count, err := cs.VPN.GetVpnGatewayByID(d.Id())
	if err != nil {
		if count == 0 {
			log.Printf(
				"[DEBUG] VPN Gateway for VPC ID %s does no longer exist", d.Get("vpc_id").(string))
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("vpc_id", v.Vpcid)
	d.Set("public_ip", v.Publicip)

	return nil
}

func resourceCosmicVPNGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cosmic.CosmicClient)

	// Create a new parameter struct
	p := cs.VPN.NewDeleteVpnGatewayParams(d.Id())

	// Delete the VPN Gateway
	_, err := cs.VPN.DeleteVpnGateway(p)
	if err != nil {
		// This is a very poor way to be told the ID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", d.Id())) {
			return nil
		}

		return fmt.Errorf("Error deleting VPN Gateway for VPC %s: %s", d.Get("vpc_id").(string), err)
	}

	return nil
}
