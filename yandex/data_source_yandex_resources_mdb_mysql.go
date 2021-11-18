package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"log"
	"regexp"
)

func dataSourceYandexResourcesMdbMySqlContent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexResourcesMdbMySqlContentRead,
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

func dataSourceYandexResourcesMdbMySqlContentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	var result []MDBResourceItem
	var clusterIds []string
	var resourcesPreset = make(map[string]MDBResourcePreset)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating network load balancer: %s", err)
	}

	clusters, err := config.sdk.MDB().MySQL().Cluster().List(ctx, &mysql.ListClustersRequest{
		FolderId: folderID,
		PageSize: 1000,
	})

	if err != nil {
		return err
	}

	for _, cluster := range clusters.Clusters {
		clusterIds = append(clusterIds, cluster.Id)
	}

	presets, err := config.sdk.MDB().MySQL().ResourcePreset().List(ctx, &mysql.ListResourcePresetsRequest{PageSize: 1000})

	if err != nil {
		return err
	}

	for _, preset := range presets.ResourcePresets {
		item := MDBResourcePreset{
			Cores:        preset.Cores,
			Memory:       preset.Memory,
			CpuPlatform:  "",
			CoreFraction: 0,
		}

		switch preset.Id {
		case "b1.nano", "b2.nano":
			item.CoreFraction = 5
		case "b1.micro", "b2.micro":
			item.CoreFraction = 20
		case "b1.medium", "b2.medium":
			item.CoreFraction = 50
		default:
			item.CoreFraction = 100
		}

		if regexp.MustCompile("^\\w+1").Match([]byte(preset.Id)) {
			item.CpuPlatform = "Intel Broadwell"
		} else if regexp.MustCompile("^\\w+2").Match([]byte(preset.Id)) {
			item.CpuPlatform = "Intel Cascade Lake"
		} else if regexp.MustCompile("^\\w+3").Match([]byte(preset.Id)) {
			item.CpuPlatform = "Intel Ice Lake"
		}

		resourcesPreset[preset.Id] = item
	}

	for _, clusterId := range clusterIds {

		cluster, err := config.sdk.MDB().MySQL().Cluster().ListHosts(ctx, &mysql.ListClusterHostsRequest{
			ClusterId: clusterId,
			PageSize:  1000,
		})

		if err != nil {
			return err
		}

		for _, host := range cluster.Hosts {
			preset := resourcesPreset[host.Resources.ResourcePresetId]

			resourceItem := MDBResourceItem{
				CpuPlatform:  preset.CpuPlatform,
				CoreFraction: preset.CoreFraction,
				Cores:        preset.Cores,
				Memory:       preset.Memory,
				NetworkSSD:   0,
				NetworkHDD:   0,
			}

			switch hddType := host.Resources.DiskTypeId; hddType {
			case "network-hdd":
				resourceItem.NetworkHDD = host.Resources.DiskSize
			case "network-ssd":
				resourceItem.NetworkSSD = host.Resources.DiskSize
			default:
				log.Printf("[INFO] type %v is not implemented currently.", hddType)
			}

			result = append(result, resourceItem)
		}
	}

	d.SetId(folderID)

	var cpuResult []map[string]interface{}

	for _, platform := range cpuPlatforms {
		var filtered []MDBResourceItem
		for _, item := range result {
			if item.CpuPlatform == platform {
				filtered = append(filtered, item)
			}
		}

		var totalCores float64 = 0
		var totalMemory int64 = 0

		for _, item := range filtered {
			totalCores += (float64(item.CoreFraction) * float64(item.Cores)) / 100
			totalMemory += item.Memory
		}
		m := make(map[string]interface{})
		m["cores"] = totalCores
		m["memory"] = totalMemory
		m["platform"] = platform
		cpuResult = append(cpuResult, m)
	}

	var totalNetworkSSD int64 = 0
	var totalNetworkHDD int64 = 0

	for _, item := range result {
		totalNetworkHDD += item.NetworkHDD
		totalNetworkSSD += item.NetworkSSD
	}

	if err := d.Set("network_ssd", totalNetworkSSD); err != nil {
		return err
	}
	if err := d.Set("network_hdd", totalNetworkHDD); err != nil {
		return err
	}
	if err := d.Set("folder_id", folderID); err != nil {
		return err
	}

	return d.Set("cpu", cpuResult)
}
