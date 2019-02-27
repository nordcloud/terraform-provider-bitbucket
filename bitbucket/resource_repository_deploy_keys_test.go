package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBitbucketDeployKeys_basic(t *testing.T) {

	testUser := os.Getenv("BITBUCKET_USERNAME")
	testAccBitbucketDeployKeysConfig := fmt.Sprintf(`
		resource "bitbucket_repository" "test_repo" {
			owner = "%s"
			name = "test-repo-deploy-keys"
		}

		resource "bitbucket_deploy_key" "testkey" {
			key = "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSUGPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XAt3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/EnmZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbxNrRFi9wrf+M7Q=="
			label = "test"
			repository = "${bitbucket_repository.test_repo.id}"
		  }
	`, testUser)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketDeployKeysDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketDeployKeysConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketDeployKeysExists("bitbucket_deploy_key.testkey", "test"),
				),
			},
		},
	})
}

func testAccCheckBitbucketDeployKeysDestroy(s *terraform.State) error {
	_, ok := s.RootModule().Resources["bitbucket_deploy_key.testkey"]
	if !ok {
		return fmt.Errorf("Not found %s", "bitbucket_deploy_key.testkey")
	}
	return nil
}

func testAccCheckBitbucketDeployKeysExists(n, label string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found %s", n)
		}

		if rs.Primary.Attributes["label"] != label {
			return fmt.Errorf("Label not set")
		}

		return nil
	}
}
