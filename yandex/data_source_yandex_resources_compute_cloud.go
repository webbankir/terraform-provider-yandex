package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"log"
)

const yandexResourcesComputeCloudLoadDisksLimit = 1000
const yandexResourcesComputeCloudLoadInstancesLimit = 1000

func dataSourceYandexResourcesComputeCloudContent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexResourcesComputeCloudContentRead,
		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"network_hdd": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"network_ssd": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"platform": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cores": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexResourcesComputeCloudContentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	var cpuResult []map[string]interface{}
	var totalNetworkSSD int64 = 0
	var totalNetworkHDD int64 = 0

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating network load balancer: %s", err)
	}

	d.SetId(folderID)

	disks, err := yandexResourcesComputeCloudLoadDisks(ctx, config, folderID, "")

	log.Printf("[DEBUG] Got disks size - %v", len(disks))

	if err != nil {
		return err
	}

	for _, item := range disks {
		switch item.TypeId {
		case "network-hdd":
			totalNetworkHDD += item.Size
		case "network-ssd":
			totalNetworkSSD += item.Size
		default:
			log.Printf("[INFO] type %v is not implemented currently.", item.TypeId)
		}
	}

	log.Printf("[DEBUG] SSD total size is - %v", totalNetworkSSD)
	log.Printf("[DEBUG] HDD total size is - %v", totalNetworkHDD)

	if err := d.Set("network_ssd", totalNetworkSSD); err != nil {
		return err
	}
	if err := d.Set("network_hdd", totalNetworkHDD); err != nil {
		return err
	}

	instances, err := yandexResourcesComputeCloudLoadInstances(ctx, config, folderID, "")

	if err != nil {
		return err
	}

	for _, platformId := range []string{"standard-v1", "standard-v2", "standard-v3"} {
		var filtered []*compute.Instance
		for _, item := range instances {
			if item.PlatformId == platformId {
				filtered = append(filtered, item)
			}
		}

		var totalCores float64 = 0
		var totalMemory int64 = 0

		for _, item := range filtered {
			totalCores += (float64(item.Resources.Cores) * float64(item.Resources.CoreFraction)) / 100
			totalMemory += item.Resources.Memory
		}

		m := make(map[string]interface{})
		m["cores"] = totalCores
		m["memory"] = totalMemory

		switch platformId {
		case "standard-v1":
			m["platform"] = "Intel Broadwell"
		case "standard-v2":
			m["platform"] = "Intel Cascade Lake"
		case "standard-v3":
			m["platform"] = "Intel Ice Lake"
		default:
			log.Printf("[INFO] platform %v is not supported here", platformId)
		}

		cpuResult = append(cpuResult, m)
	}

	if err := d.Set("folder_id", folderID); err != nil {
		return err
	}
	return d.Set("cpu", cpuResult)
}

func yandexResourcesComputeCloudLoadInstances(ctx context.Context, config *Config, folderId string, nextPageToken string) ([]*compute.Instance, error) {
	var instances []*compute.Instance
	disk, err := config.sdk.Compute().Instance().List(ctx, &compute.ListInstancesRequest{
		PageSize:  yandexResourcesComputeCloudLoadInstancesLimit,
		FolderId:  folderId,
		PageToken: nextPageToken,
	})

	if err != nil {
		return instances, err
	}

	for _, item := range disk.Instances {
		if item.Status == compute.Instance_RUNNING {
			instances = append(instances, item)
		}
	}

	if len(disk.Instances) == yandexResourcesComputeCloudLoadDisksLimit {
		data, err := yandexResourcesComputeCloudLoadInstances(ctx, config, folderId, disk.NextPageToken)

		if len(data) > 0 {
			return append(instances, data...), nil
		}
		if err != nil {
			return instances, err
		}
	}

	return instances, nil
}

func yandexResourcesComputeCloudLoadDisks(ctx context.Context, config *Config, folderId string, nextPageToken string) ([]*compute.Disk, error) {
	var disks []*compute.Disk
	disk, err := config.sdk.Compute().Disk().List(ctx, &compute.ListDisksRequest{
		PageSize:  yandexResourcesComputeCloudLoadDisksLimit,
		FolderId:  folderId,
		PageToken: nextPageToken,
	})

	if err != nil {
		return disks, err
	}

	for _, item := range disk.Disks {
		disks = append(disks, item)
	}

	if len(disk.Disks) == yandexResourcesComputeCloudLoadDisksLimit {
		data, err := yandexResourcesComputeCloudLoadDisks(ctx, config, folderId, disk.NextPageToken)

		if len(data) > 0 {
			return append(disks, data...), nil
		}
		if err != nil {
			return disks, err
		}
	}

	return disks, nil
}
