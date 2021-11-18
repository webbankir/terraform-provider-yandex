package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"log"
)


func dataSourceYandexLockBoxSecretPayload() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexLockBoxSecretPayloadRead,
		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type: schema.TypeString,
				Required: true,
			},

			"key": {
				Type: schema.TypeString,
				Optional: true,
			},

			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"values": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

		},
	}
}

func dataSourceYandexLockBoxSecretPayloadRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	secretId := d.Get("secret_id").(string)
	key, exists  := d.GetOkExists("key")

	d.SetId(secretId)

	log.Printf("[DEBUG] secret_id=> '%v' , key => '%v'\n\n", secretId, key)

	payload, err := config.sdk.LockboxPayload().Payload().Get(ctx, &lockbox.GetPayloadRequest{SecretId: secretId})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Secret %q", d.Id()))
	}

	values := make(map[string]string)

	for _, v := range payload.Entries {
		values[v.Key] = v.GetTextValue()
		if exists && v.Key == key.(string) {
			log.Printf("[DEBUG] set => '%v'\n", v.GetTextValue())

			if err := d.Set("value", v.GetTextValue()); err != nil {
				return err
			}
		}
	}

	if err := d.Set("secret_id", secretId); err != nil {
		return err
	}

	if err := d.Set("key", key); err != nil {
		return err
	}

	return d.Set("values", values)
}
