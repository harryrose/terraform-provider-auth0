package auth0

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccDataClient(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"auth0": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataClientConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_id", "my_client_id"),
				),
			},
		},
	})
}

const testAccDataClientConfig = `
provider "auth0" {
	client_id = "provider_client_id"
	client_secret = "provider_client_secret"
	domain = "provider_domain"
}

data "auth0_client" "my_client" {
  client_id = "my_client_id"
}
`