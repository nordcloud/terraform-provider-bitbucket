package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
)

// DeployKey structure for handling key info
type DeployKey struct {
	ID    int    `json:"id,omitempty"`
	Key   string `json:"key,omitempty"`
	Label string `json:"label,omitempty"`
}

func resourceDeployKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeployKeyCreate,
		Update: resourceDeployKeyUpdate,
		Read:   resourceDeployKeyRead,
		Delete: resourceDeployKeyDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"label": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func newDeployKeyFromResource(d *schema.ResourceData) *DeployKey {
	dk := &DeployKey{
		ID:    d.Get("id").(int),
		Key:   d.Get("key").(string),
		Label: d.Get("label").(string),
	}
	return dk
}

func resourceDeployKeyCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)
	dk := newDeployKeyFromResource(d)

	bytedata, err := json.Marshal(dk)

	if err != nil {
		return err
	}

	req, err := client.Post(fmt.Sprintf("2.0/repositories/%s/deploy-keys",
		d.Get("repository").(string),
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	body, readerr := ioutil.ReadAll(req.Body)
	if readerr != nil {
		return readerr
	}

	decodeerr := json.Unmarshal(body, &dk)
	if decodeerr != nil {
		return decodeerr
	}

	d.SetId(strconv.Itoa(dk.ID))

	return resourceDeployKeyRead(d, m)
}

func resourceDeployKeyRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)
	dkReq, _ := client.Get(fmt.Sprintf("2.0/repositories/%s/deploy-keys/%s",
		d.Get("repository").(string),
		d.Id(),
	))

	if dkReq.StatusCode == 200 {
		var dk DeployKey
		body, readerr := ioutil.ReadAll(dkReq.Body)
		if readerr != nil {
			return readerr
		}

		decodeerr := json.Unmarshal(body, &dk)
		if decodeerr != nil {
			return decodeerr
		}

		d.Set("id", dk.ID)
		d.Set("key", dk.Key)
		d.Set("label", dk.Label)
	}

	if dkReq.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceDeployKeyUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)
	dk := newDeployKeyFromResource(d)

	bytedata, err := json.Marshal(dk)

	if err != nil {
		return err
	}

	req, err := client.Delete(fmt.Sprintf("2.0/repositories/%s/deploy-keys/%s",
		d.Get("repository").(string),
		d.Id(),
	))

	if req.StatusCode != 204 {
		log.Printf("[ERROR] Could not delete the key: %s", d.Id())
		return nil
	}

	req, err = client.Post(fmt.Sprintf("2.0/repositories/%s/deploy-keys",
		d.Get("repository").(string),
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	if req.StatusCode != 200 {
		return nil
	}

	return resourceDeployKeyRead(d, m)
}

func resourceDeployKeyDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)
	_, err := client.Delete(fmt.Sprintf("2.0/repositories/%s/deploy-keys/%s",
		d.Get("repository").(string),
		d.Id(),
	))
	return err
}
