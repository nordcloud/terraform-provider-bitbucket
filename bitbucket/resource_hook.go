package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/hashicorp/terraform/helper/schema"
)

type Hook struct {
	Uuid                 string   `json:"uuid,omitempty"`
	Url                  string   `json:"url,omitempty"`
	Description          string   `json:"description,omitempty"`
	Active               bool     `json:"active,omitempty"`
	SkipCertVerification bool     `json:"skip_cert_verification,omitempty"`
	Events               []string `json:"events,omitempty"`
}

func resourceHook() *schema.Resource {
	return &schema.Resource{
		Create: resourceHookCreate,
		Read:   resourceHookRead,
		Update: resourceHookUpdate,
		Delete: resourceHookDelete,
		Exists: resourceHookExists,

		Schema: map[string]*schema.Schema{
			"owner": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"events": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"skip_cert_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func createHook(d *schema.ResourceData) *Hook {

	events := make([]string, 0, len(d.Get("events").(*schema.Set).List()))

	for _, item := range d.Get("events").(*schema.Set).List() {
		events = append(events, item.(string))
	}

	return &Hook{
		Url:                  d.Get("url").(string),
		Description:          d.Get("description").(string),
		Active:               d.Get("active").(bool),
		SkipCertVerification: d.Get("skip_cert_verification").(bool),
		Events:               events,
	}
}

func resourceHookCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)
	hook := createHook(d)

	payload, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	hook_req, err := client.Post(fmt.Sprintf("2.0/repositories/%s/%s/hooks",
		d.Get("owner").(string),
		d.Get("repository").(string),
	), bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	body, readerr := ioutil.ReadAll(hook_req.Body)
	if readerr != nil {
		return readerr
	}

	decodeerr := json.Unmarshal(body, &hook)
	if decodeerr != nil {
		return decodeerr
	}

	d.SetId(hook.Uuid)

	return resourceHookRead(d, m)
}
func resourceHookRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)

	hook_req, _ := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/hooks/%s",
		d.Get("owner").(string),
		d.Get("repository").(string),
		url.PathEscape(d.Id()),
	))

	log.Printf("ID: %s", url.PathEscape(d.Id()))

	if hook_req.StatusCode == 200 {
		var hook Hook

		body, readerr := ioutil.ReadAll(hook_req.Body)
		if readerr != nil {
			return readerr
		}

		decodeerr := json.Unmarshal(body, &hook)
		if decodeerr != nil {
			return decodeerr
		}

		d.Set("uuid", hook.Uuid)
		d.Set("description", hook.Description)
		d.Set("active", hook.Active)
		d.Set("url", hook.Url)
		d.Set("skip_cert_verification", hook.SkipCertVerification)

		eventsList := make([]string, 0, len(hook.Events))

		for _, event := range hook.Events {
			eventsList = append(eventsList, event)
		}

		d.Set("events", eventsList)
	}

	return nil
}

func resourceHookUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)
	hook := createHook(d)
	payload, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	_, err = client.Put(fmt.Sprintf("2.0/repositories/%s/%s/hooks/%s",
		d.Get("owner").(string),
		d.Get("repository").(string),
		url.PathEscape(d.Id()),
	), bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	return resourceHookRead(d, m)
}

func resourceHookExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*BitbucketClient)
	if _, okay := d.GetOk("uuid"); okay {
		hook_req, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/hooks/%s",
			d.Get("owner").(string),
			d.Get("repository").(string),
			url.PathEscape(d.Id()),
		))

		if err != nil {
			log.Printf("[DEBUG] Req: %+v, Err: %+v", hook_req, err)
			// If the hook was not found, we get the message "is not a valid hook".
			// Return nil so we can show that the hook is gone.
			if hook_req.StatusCode == 404 {
				return false, nil
			}

			panic(err)
		}

		if hook_req.StatusCode != 200 {
			return false, err
		}

		return true, nil
	}

	return false, nil

}

func resourceHookDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*BitbucketClient)
	_, err := client.Delete(fmt.Sprintf("2.0/repositories/%s/%s/hooks/%s",
		d.Get("owner").(string),
		d.Get("repository").(string),
		url.PathEscape(d.Id()),
	))

	return err

}
