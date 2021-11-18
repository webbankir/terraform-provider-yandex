package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/version"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/mutexkv"
)

const (
	defaultMaxRetries      = 5
	defaultEndpoint        = "api.cloud.yandex.net:443"
	defaultStorageEndpoint = "storage.yandexcloud.net"
	defaultYMQEndpoint     = "message-queue.api.cloud.yandex.net"
)

// Global MutexKV
var mutexKV = mutexkv.NewMutexKV()

func Provider() *schema.Provider {
	return provider(false)
}

func emptyFolderProvider() *schema.Provider {
	return provider(true)
}

func provider(emptyFolder bool) *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_ENDPOINT", defaultEndpoint),
				Description: descriptions["endpoint"],
			},
			"folder_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_FOLDER_ID", nil),
				Description: descriptions["folder_id"],
			},
			"cloud_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_CLOUD_ID", nil),
				Description: descriptions["cloud_id"],
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_ZONE", nil),
				Description: descriptions["zone"],
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_TOKEN", nil),
				Description: descriptions["token"],
			},
			"service_account_key_file": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("YC_SERVICE_ACCOUNT_KEY_FILE", nil),
				Description:   descriptions["service_account_key_file"],
				ConflictsWith: []string{"token"},
				ValidateFunc:  validateSAKey,
			},
			"storage_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_STORAGE_ENDPOINT_URL", defaultStorageEndpoint),
				Description: descriptions["storage_endpoint"],
			},
			"storage_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_STORAGE_ACCESS_KEY", nil),
				Description: descriptions["storage_access_key"],
			},
			"storage_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_STORAGE_SECRET_KEY", nil),
				Description: descriptions["storage_secret_key"],
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_INSECURE", false),
				Description: descriptions["insecure"],
			},
			"plaintext": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_PLAINTEXT", false),
				Description: descriptions["plaintext"],
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     defaultMaxRetries,
				Description: descriptions["max_retries"],
			},
			"ymq_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_MESSAGE_QUEUE_ENDPOINT", defaultYMQEndpoint),
				Description: descriptions["ymq_endpoint"],
			},
			"ymq_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_MESSAGE_QUEUE_ACCESS_KEY", nil),
				Description: descriptions["ymq_access_key"],
			},
			"ymq_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("YC_MESSAGE_QUEUE_SECRET_KEY", nil),
				Description: descriptions["ymq_secret_key"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"yandex_alb_backend_group":            dataSourceYandexALBBackendGroup(),
			"yandex_alb_http_router":              dataSourceYandexALBHTTPRouter(),
			"yandex_alb_load_balancer":            dataSourceYandexALBLoadBalancer(),
			"yandex_alb_target_group":             dataSourceYandexALBTargetGroup(),
			"yandex_alb_virtual_host":             dataSourceYandexALBVirtualHost(),
			"yandex_api_gateway":                  dataSourceYandexApiGateway(),
			"yandex_client_config":                dataSourceYandexClientConfig(),
			"yandex_container_registry":           dataSourceYandexContainerRegistry(),
			"yandex_container_repository":         dataSourceYandexContainerRepository(),
			"yandex_compute_disk":                 dataSourceYandexComputeDisk(),
			"yandex_compute_disk_placement_group": dataSourceYandexComputeDiskPlacementGroup(),
			"yandex_compute_image":                dataSourceYandexComputeImage(),
			"yandex_compute_instance":             dataSourceYandexComputeInstance(),
			"yandex_compute_instance_group":       dataSourceYandexComputeInstanceGroup(),
			"yandex_compute_placement_group":      dataSourceYandexComputePlacementGroup(),
			"yandex_compute_snapshot":             dataSourceYandexComputeSnapshot(),
			"yandex_dataproc_cluster":             dataSourceYandexDataprocCluster(),
			"yandex_dns_zone":                     dataSourceYandexDnsZone(),
			"yandex_function":                     dataSourceYandexFunction(),
			"yandex_function_scaling_policy":      dataSourceYandexFunctionScalingPolicy(),
			"yandex_function_trigger":             dataSourceYandexFunctionTrigger(),
			"yandex_iam_policy":                   dataSourceYandexIAMPolicy(),
			"yandex_iam_role":                     dataSourceYandexIAMRole(),
			"yandex_iam_service_account":          dataSourceYandexIAMServiceAccount(),
			"yandex_iam_user":                     dataSourceYandexIAMUser(),
			"yandex_iot_core_device":              dataSourceYandexIoTCoreDevice(),
			"yandex_iot_core_registry":            dataSourceYandexIoTCoreRegistry(),
			"yandex_kubernetes_cluster":           dataSourceYandexKubernetesCluster(),
			"yandex_kubernetes_node_group":        dataSourceYandexKubernetesNodeGroup(),
			"yandex_lb_network_load_balancer":     dataSourceYandexLBNetworkLoadBalancer(),
			"yandex_lb_target_group":              dataSourceYandexLBTargetGroup(),
			"yandex_logging_group":                dataSourceYandexLoggingGroup(),
			"yandex_mdb_clickhouse_cluster":       dataSourceYandexMDBClickHouseCluster(),
			"yandex_mdb_mongodb_cluster":          dataSourceYandexMDBMongodbCluster(),
			"yandex_mdb_mysql_cluster":            dataSourceYandexMDBMySQLCluster(),
			"yandex_mdb_sqlserver_cluster":        dataSourceYandexMDBSQLServerCluster(),
			"yandex_mdb_greenplum_cluster":        dataSourceYandexMDBGreenplumCluster(),
			"yandex_mdb_postgresql_cluster":       dataSourceYandexMDBPostgreSQLCluster(),
			"yandex_mdb_redis_cluster":            dataSourceYandexMDBRedisCluster(),
			"yandex_mdb_kafka_cluster":            dataSourceYandexMDBKafkaCluster(),
			"yandex_mdb_kafka_topic":              dataSourceYandexMDBKafkaTopic(),
			"yandex_mdb_elasticsearch_cluster":    dataSourceYandexMDBElasticsearchCluster(),
			"yandex_message_queue":                dataSourceYandexMessageQueue(),
			"yandex_resourcemanager_cloud":        dataSourceYandexResourceManagerCloud(),
			"yandex_resourcemanager_folder":       dataSourceYandexResourceManagerFolder(),
			"yandex_vpc_address":                  dataSourceYandexVPCAddress(),
			"yandex_vpc_network":                  dataSourceYandexVPCNetwork(),
			"yandex_vpc_route_table":              dataSourceYandexVPCRouteTable(),
			"yandex_vpc_security_group":           dataSourceYandexVPCSecurityGroup(),
			"yandex_vpc_security_group_rule":      dataSourceYandexVPCSecurityGroupRule(),
			"yandex_vpc_subnet":                   dataSourceYandexVPCSubnet(),
			"yandex_ydb_database_dedicated":       dataSourceYandexYDBDatabaseDedicated(),
			"yandex_ydb_database_serverless":      dataSourceYandexYDBDatabaseServerless(),

			"yandex_lockbox_secret_payload":       dataSourceYandexLockBoxSecretPayload(),
			"yandex_certificate_manager_list":     dataSourceYandexCertificateManagerList(),
			"yandex_certificate_manager_content":  dataSourceYandexCertificateManagerContent(),
			"yandex_billing_account":              dataSourceYandexBillingAccountContent(),
			"yandex_resource_mdb_mysql":           dataSourceYandexResourcesMdbMySqlContent(),
			"yandex_resource_mdb_postgresql":      dataSourceYandexResourcesMdbPostgreSqlContent(),
			"yandex_resource_mdb_mongodb":         dataSourceYandexResourcesMdbMongoDbContent(),
			"yandex_resource_mdb_redis":           dataSourceYandexResourcesMdbRedisContent(),
			"yandex_resource_compute_cloud":       dataSourceYandexResourcesComputeCloudContent(),

		},

		ResourcesMap: map[string]*schema.Resource{
			"yandex_alb_backend_group":                            resourceYandexALBBackendGroup(),
			"yandex_alb_http_router":                              resourceYandexALBHTTPRouter(),
			"yandex_alb_load_balancer":                            resourceYandexALBLoadBalancer(),
			"yandex_alb_target_group":                             resourceYandexALBTargetGroup(),
			"yandex_alb_virtual_host":                             addPassthroughImport(withALBVirtualHostID(resourceYandexALBVirtualHost())),
			"yandex_api_gateway":                                  resourceYandexApiGateway(),
			"yandex_container_registry":                           resourceYandexContainerRegistry(),
			"yandex_container_registry_iam_binding":               resourceYandexContainerRegistryIAMBinding(),
			"yandex_container_repository":                         resourceYandexContainerRepository(),
			"yandex_container_repository_iam_binding":             resourceYandexContainerRepositoryIAMBinding(),
			"yandex_compute_disk":                                 resourceYandexComputeDisk(),
			"yandex_compute_disk_placement_group":                 resourceYandexComputeDiskPlacementGroup(),
			"yandex_compute_image":                                resourceYandexComputeImage(),
			"yandex_compute_instance":                             resourceYandexComputeInstance(),
			"yandex_compute_instance_group":                       resourceYandexComputeInstanceGroup(),
			"yandex_compute_snapshot":                             resourceYandexComputeSnapshot(),
			"yandex_compute_placement_group":                      resourceYandexComputePlacementGroup(),
			"yandex_dataproc_cluster":                             resourceYandexDataprocCluster(),
			"yandex_dns_recordset":                                resourceYandexDnsRecordSet(),
			"yandex_dns_zone":                                     resourceYandexDnsZone(),
			"yandex_function_iam_binding":                         resourceYandexFunctionIAMBinding(),
			"yandex_function":                                     resourceYandexFunction(),
			"yandex_function_scaling_policy":                      resourceYandexFunctionScalingPolicy(),
			"yandex_function_trigger":                             resourceYandexFunctionTrigger(),
			"yandex_iam_service_account":                          resourceYandexIAMServiceAccount(),
			"yandex_iam_service_account_api_key":                  resourceYandexIAMServiceAccountAPIKey(),
			"yandex_iam_service_account_iam_binding":              resourceYandexIAMServiceAccountIAMBinding(),
			"yandex_iam_service_account_iam_member":               resourceYandexIAMServiceAccountIAMMember(),
			"yandex_iam_service_account_iam_policy":               resourceYandexIAMServiceAccountIAMPolicy(),
			"yandex_iam_service_account_key":                      resourceYandexIAMServiceAccountKey(),
			"yandex_iam_service_account_static_access_key":        resourceYandexIAMServiceAccountStaticAccessKey(),
			"yandex_iot_core_device":                              resourceYandexIoTCoreDevice(),
			"yandex_iot_core_registry":                            resourceYandexIoTCoreRegistry(),
			"yandex_kms_symmetric_key_iam_binding":                resourceYandexKMSSymmetricKeyIAMBinding(),
			"yandex_kms_symmetric_key":                            resourceYandexKMSSymmetricKeyKey(),
			"yandex_kms_secret_ciphertext":                        resourceYandexKMSSecretCiphertext(),
			"yandex_kubernetes_cluster":                           resourceYandexKubernetesCluster(),
			"yandex_kubernetes_node_group":                        resourceYandexKubernetesNodeGroup(),
			"yandex_lb_network_load_balancer":                     resourceYandexLBNetworkLoadBalancer(),
			"yandex_lb_target_group":                              resourceYandexLBTargetGroup(),
			"yandex_logging_group":                                resourceYandexLoggingGroup(),
			"yandex_mdb_clickhouse_cluster":                       resourceYandexMDBClickHouseCluster(),
			"yandex_mdb_mongodb_cluster":                          resourceYandexMDBMongodbCluster(),
			"yandex_mdb_mysql_cluster":                            resourceYandexMDBMySQLCluster(),
			"yandex_mdb_sqlserver_cluster":                        resourceYandexMDBSQLServerCluster(),
			"yandex_mdb_greenplum_cluster":                        resourceYandexMDBGreenplumCluster(),
			"yandex_mdb_postgresql_cluster":                       resourceYandexMDBPostgreSQLCluster(),
			"yandex_mdb_redis_cluster":                            resourceYandexMDBRedisCluster(),
			"yandex_mdb_kafka_cluster":                            resourceYandexMDBKafkaCluster(),
			"yandex_mdb_kafka_topic":                              resourceYandexMDBKafkaTopic(),
			"yandex_mdb_elasticsearch_cluster":                    resourceYandexMDBElasticsearchCluster(),
			"yandex_message_queue":                                resourceYandexMessageQueue(),
			"yandex_organizationmanager_organization_iam_member":  resourceYandexOrganizationManagerOrganizationIAMMember(),
			"yandex_organizationmanager_organization_iam_binding": resourceYandexOrganizationManagerOrganizationIAMBinding(),
			"yandex_resourcemanager_cloud_iam_binding":            resourceYandexResourceManagerCloudIAMBinding(),
			"yandex_resourcemanager_cloud_iam_member":             resourceYandexResourceManagerCloudIAMMember(),
			"yandex_resourcemanager_folder_iam_binding":           resourceYandexResourceManagerFolderIAMBinding(),
			"yandex_resourcemanager_folder_iam_member":            resourceYandexResourceManagerFolderIAMMember(),
			"yandex_resourcemanager_folder_iam_policy":            resourceYandexResourceManagerFolderIAMPolicy(),
			"yandex_resourcemanager_folder":                       resourceYandexResourceManagerFolder(),
			"yandex_storage_bucket":                               resourceYandexStorageBucket(),
			"yandex_storage_object":                               resourceYandexStorageObject(),
			"yandex_vpc_address":                                  resourceYandexVPCAddress(),
			"yandex_vpc_network":                                  resourceYandexVPCNetwork(),
			"yandex_vpc_route_table":                              resourceYandexVPCRouteTable(),
			"yandex_vpc_security_group":                           resourceYandexVPCSecurityGroup(),
			"yandex_vpc_default_security_group":                   resourceYandexVPCDefaultSecurityGroup(),
			"yandex_vpc_security_group_rule":                      resourceYandexVpcSecurityGroupRule(),
			"yandex_vpc_subnet":                                   resourceYandexVPCSubnet(),
			"yandex_ydb_database_dedicated":                       resourceYandexYDBDatabaseDedicated(),
			"yandex_ydb_database_serverless":                      resourceYandexYDBDatabaseServerless(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(ctx, d, provider, emptyFolder)
	}

	return provider
}

func addPassthroughImport(r *schema.Resource) *schema.Resource {
	r.Importer = &schema.ResourceImporter{
		State: schema.ImportStatePassthrough,
	}
	return r
}

type crudFunc = func(d *schema.ResourceData, meta interface{}) error

func withALBVirtualHostID(r *schema.Resource) *schema.Resource {
	r.Read = wrapParseVirtualHostID(r.Read)
	r.Update = wrapParseVirtualHostID(r.Update)
	r.Delete = wrapParseVirtualHostID(r.Delete)
	return r
}

func wrapParseVirtualHostID(f crudFunc) crudFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		attrs := strings.Split(d.Id(), "/")
		if len(attrs) < 2 {
			return fmt.Errorf("error reading virtual_host, wrong id: %q", d.Id())
		}
		if err := d.Set("http_router_id", attrs[0]); err != nil {
			return err
		}
		if err := d.Set("name", attrs[1]); err != nil {
			return err
		}
		return f(d, meta)
	}
}

var descriptions = map[string]string{
	"endpoint": "The API endpoint for Yandex.Cloud SDK client.",

	"folder_id": "The default folder ID where resources will be placed.",

	"cloud_id": "ID of Yandex.Cloud tenant.",

	"zone": "The zone where operations will take place. Examples\n" +
		"are ru-central1-a, ru-central2-c, etc.",

	"token": "The access token for API operations.",

	"service_account_key_file": "Either the path to or the contents of a Service Account key file in JSON format.",

	"insecure": "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted," +
		"default value is `false`.",

	"plaintext": "Disable use of TLS. Default value is `false`.",

	"max_retries": "The maximum number of times an API request is being executed. \n" +
		"If the API request still fails, an error is thrown.",

	"storage_endpoint": "Yandex.Cloud storage service endpoint. Default is \n" + defaultStorageEndpoint,

	"storage_access_key": "Yandex.Cloud storage service access key. \n" +
		"Used when a storage data/resource doesn't have an access key explicitly specified.",

	"storage_secret_key": "Yandex.Cloud storage service secret key. \n" +
		"Used when a storage data/resource doesn't have a secret key explicitly specified.",

	"ymq_endpoint": "Yandex.Cloud Message Queue service endpoint. Default is \n" + defaultYMQEndpoint,

	"ymq_access_key": "Yandex.Cloud Message Queue service access key. \n" +
		"Used when a message queue resource doesn't have an access key explicitly specified.",

	"ymq_secret_key": "Yandex.Cloud Message Queue service secret key. \n" +
		"Used when a message queue resource doesn't have a secret key explicitly specified.",
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, p *schema.Provider, emptyFolder bool) (interface{}, diag.Diagnostics) {
	//return func(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token:                          d.Get("token").(string),
		ServiceAccountKeyFileOrContent: d.Get("service_account_key_file").(string),
		Zone:                           d.Get("zone").(string),
		FolderID:                       d.Get("folder_id").(string),
		CloudID:                        d.Get("cloud_id").(string),
		Endpoint:                       d.Get("endpoint").(string),
		Plaintext:                      d.Get("plaintext").(bool),
		Insecure:                       d.Get("insecure").(bool),
		MaxRetries:                     d.Get("max_retries").(int),
		StorageEndpoint:                d.Get("storage_endpoint").(string),
		StorageAccessKey:               d.Get("storage_access_key").(string),
		StorageSecretKey:               d.Get("storage_secret_key").(string),
		YMQEndpoint:                    d.Get("ymq_endpoint").(string),
		YMQAccessKey:                   d.Get("ymq_access_key").(string),
		YMQSecretKey:                   d.Get("ymq_secret_key").(string),
		userAgent:                      p.UserAgent("terraform-provider-yandex", version.ProviderVersion),
	}

	if emptyFolder {
		config.FolderID = ""
	}

	stopCtx, ok := schema.StopContext(ctx)
	if !ok {
		stopCtx = ctx
	}
	terraformVersion := p.TerraformVersion
	if terraformVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		terraformVersion = "0.11+compatible"
	}

	if err := config.initAndValidate(stopCtx, terraformVersion, false); err != nil {
		return nil, diag.FromErr(err)
	}

	return &config, nil

}

func validateSAKey(v interface{}, k string) (warnings []string, errors []error) {
	if v == nil || v.(string) == "" {
		return
	}

	saKey := v.(string)
	// if this is a path to file and we can stat it, assume it's ok
	if _, err := os.Stat(saKey); err == nil {
		return
	}

	// else check for a valid json data value
	var f map[string]interface{}
	if err := json.Unmarshal([]byte(saKey), &f); err != nil {
		errors = append(errors, fmt.Errorf("JSON in %q are not valid: %s", saKey, err))
	}

	return
}
