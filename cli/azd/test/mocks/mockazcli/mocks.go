package mockazcli

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/azure/azure-dev/cli/azd/pkg/azapi"
	"github.com/azure/azure-dev/cli/azd/pkg/azsdk"
	"github.com/azure/azure-dev/cli/azd/pkg/tools/azcli"
	"github.com/azure/azure-dev/cli/azd/test/mocks"
	"github.com/azure/azure-dev/cli/azd/test/mocks/mockaccount"
)

// NewAzCliFromMockContext creates a new instance of AzCli, configured to use the credential and pipeline from the
// provided mock context.
func NewAzCliFromMockContext(mockContext *mocks.MockContext) azcli.AzCli {
	// nolint:lll
	deletedServicesClient, _ := armapimanagement.NewDeletedServicesClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	serviceClient, _ := armapimanagement.NewServiceClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	configurationStoresClient, _ := armappconfiguration.NewConfigurationStoresClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	accountsClient, _ := armcognitiveservices.NewAccountsClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	deletedAccountsClient, _ := armcognitiveservices.NewDeletedAccountsClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	vaultsClient, _ := armkeyvault.NewVaultsClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	managedHsmsClient, _ := armkeyvault.NewManagedHsmsClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	client, _ := armresources.NewClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	resourceGroupsClient, _ := armresources.NewResourceGroupsClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	staticSitesClient, _ := armappservice.NewStaticSitesClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	webAppsClient, _ := armappservice.NewWebAppsClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	zipDeployClient, _ := azsdk.NewZipDeployClient("SUBSCRIPTION_ID", mockContext.Credentials, mockContext.ArmClientOptions)
	// nolint:end

	return azcli.NewAzCli(
		mockaccount.SubscriptionCredentialProviderFunc(func(_ context.Context, _ string) (azcore.TokenCredential, error) {
			return mockContext.Credentials, nil
		}),
		mockContext.HttpClient,
		azcli.NewAzCliArgs{},
		deletedServicesClient,
		serviceClient,
		configurationStoresClient,
		accountsClient,
		deletedAccountsClient,
		vaultsClient,
		managedHsmsClient,
		client,
		resourceGroupsClient,
		staticSitesClient,
		webAppsClient,
		zipDeployClient,
	)
}

func NewDeploymentOperationsServiceFromMockContext(
	mockContext *mocks.MockContext) azapi.DeploymentOperations {
	client, _ := armresources.NewDeploymentOperationsClient(
		"SUBSCRIPTION_ID", // TODO: this probably needs to be mocked
		mockContext.Credentials,
		mockContext.ArmClientOptions,
	)

	return azapi.NewDeploymentOperations(client)
}

func NewDeploymentsServiceFromMockContext(
	mockContext *mocks.MockContext) azapi.Deployments {
	client, _ := armresources.NewDeploymentsClient(
		"SUBSCRIPTION_ID", // TODO: this probably needs to be mocked
		mockContext.Credentials,
		mockContext.ArmClientOptions,
	)
	return azapi.NewDeployments(client)
}
