package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
	"strings"
)

func dataSourceYandexCertificateManagerContent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexCertificateManagerContentRead,
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type: schema.TypeString,
				Required: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},

		},
	}
}

func dataSourceYandexCertificateManagerContentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	certificateId := d.Get("certificate_id").(string)
	d.SetId(certificateId)

	data, err := config.sdk.CertificatesData().CertificateContent().Get(ctx, &certificatemanager.GetCertificateContentRequest{CertificateId: certificateId })
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Certificate Id %q", d.Id()))
	}
	if err := d.Set("private_key", data.PrivateKey); err != nil {
		return err
	}
	if err := d.Set("certificate_chain", strings.Join(data.CertificateChain, "\n")); err != nil {
		return err
	}
	return d.Set("certificate_id", certificateId)
}