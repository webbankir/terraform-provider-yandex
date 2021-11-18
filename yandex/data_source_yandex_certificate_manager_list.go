package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
)

//	Type CertificateType `protobuf:"varint,7,opt,name=type,proto3,enum=yandex.cloud.certificatemanager.v1.CertificateType" json:"type,omitempty"`
//	// Fully qualified domain names of the certificate.
//	Status Certificate_Status `protobuf:"varint,9,opt,name=status,proto3,enum=yandex.cloud.certificatemanager.v1.Certificate_Status" json:"status,omitempty"`
//	// [Distinguished Name](https://tools.ietf.org/html/rfc1779) of the certificate authority that issued the certificate.
//	Issuer string `protobuf:"bytes,10,opt,name=issuer,proto3" json:"issuer,omitempty"`
//	// [Distinguished Name](https://tools.ietf.org/html/rfc1779) of the entity that is associated with the public key contained in the certificate.
//	Subject string `protobuf:"bytes,11,opt,name=subject,proto3" json:"subject,omitempty"`
//	// Serial number of the certificate.
//	Serial string `protobuf:"bytes,12,opt,name=serial,proto3" json:"serial,omitempty"`
//	// Time when the certificate is updated.
//	UpdatedAt *timestamp.Timestamp `protobuf:"bytes,13,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
//	// Time when the certificate is issued.
//	IssuedAt *timestamp.Timestamp `protobuf:"bytes,14,opt,name=issued_at,json=issuedAt,proto3" json:"issued_at,omitempty"`
//	// Time after which the certificate is not valid.
//	NotAfter *timestamp.Timestamp `protobuf:"bytes,15,opt,name=not_after,json=notAfter,proto3" json:"not_after,omitempty"`
//	// Time before which the certificate is not valid.
//	NotBefore *timestamp.Timestamp `protobuf:"bytes,16,opt,name=not_before,json=notBefore,proto3" json:"not_before,omitempty"`
//	// Domains validation challenges of the certificate. Used only for managed certificates.
//	Challenges []*Challenge `protobuf:"bytes,17,rep,name=challenges,proto3" json:"challenges,omitempty"`
type M map[string]interface{}

func dataSourceYandexCertificateManagerList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexCertificateManagerListRead,
		Schema: map[string]*schema.Schema{
			"values": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"folder_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"labels": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},

						"domains": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexCertificateManagerListRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	d.SetId(config.FolderID)

	list, err := config.sdk.Certificates().Certificate().List(ctx, &certificatemanager.ListCertificatesRequest{FolderId: config.FolderID, PageSize: 1000})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Secret %q", d.Id()))
	}

	var values []M

	for _, v := range list.Certificates {

		values = append(values, M{
			"id":          v.Id,
			"folder_id":   v.FolderId,
			"name":        v.Name,
			"description": v.Description,
			"labels":      v.Labels,
			"domains":     v.Domains,
			"status":      certificatemanager.Certificate_Status_name[int32(v.Status)],
		})
	}

	return d.Set("values", values)
}
