// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package azcli

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/azure/azure-dev/cli/azd/pkg/azsdk"
	"github.com/azure/azure-dev/cli/azd/pkg/convert"
	"github.com/azure/azure-dev/cli/azd/test/mocks"
	"github.com/azure/azure-dev/cli/azd/test/mocks/mockaccount"
	"github.com/stretchr/testify/require"
)

func TestAZCLIWithUserAgent(t *testing.T) {
	mockContext := mocks.NewMockContext(context.Background())
	mockContext.HttpClient.When(func(request *http.Request) bool {
		return request.Method == http.MethodGet && request.URL.Path == "/RESOURCE_ID"
	}).RespondFn(func(request *http.Request) (*http.Response, error) {
		response := armresources.ClientGetByIDResponse{
			GenericResource: armresources.GenericResource{
				ID:       convert.RefOf("RESOURCE_ID"),
				Kind:     convert.RefOf("RESOURCE_KIND"),
				Name:     convert.RefOf("RESOURCE_NAME"),
				Type:     convert.RefOf("RESOURCE_TYPE"),
				Location: convert.RefOf("RESOURCE_LOCATION"),
			},
		}

		return mocks.CreateHttpResponseWithBody(request, http.StatusOK, response)
	})

	var rawResponse *http.Response
	ctx := runtime.WithCaptureResponse(*mockContext.Context, &rawResponse)

	azCli := newAzCliFromMockContext(mockContext)
	// We don't care about the actual response or if an error occurred
	// Any API call that leverages the Go SDK is fine
	_, _ = azCli.GetResource(ctx, "SUBSCRIPTION_ID", "RESOURCE_ID", "API_VERSION")

	userAgent, ok := rawResponse.Request.Header["User-Agent"]
	if !ok {
		require.Fail(t, "missing User-Agent header")
	}

	require.Contains(t, userAgent[0], "azsdk-go")
	require.Contains(t, userAgent[0], "azdev")
}

func Test_AzSdk_User_Agent_Policy(t *testing.T) {
	mockContext := mocks.NewMockContext(context.Background())
	mockContext.HttpClient.When(func(request *http.Request) bool {
		return request.Method == http.MethodGet && request.URL.Path == "/RESOURCE_ID"
	}).RespondFn(func(request *http.Request) (*http.Response, error) {
		response := armresources.ClientGetByIDResponse{
			GenericResource: armresources.GenericResource{
				ID:       convert.RefOf("RESOURCE_ID"),
				Kind:     convert.RefOf("RESOURCE_KIND"),
				Name:     convert.RefOf("RESOURCE_NAME"),
				Type:     convert.RefOf("RESOURCE_TYPE"),
				Location: convert.RefOf("RESOURCE_LOCATION"),
			},
		}

		return mocks.CreateHttpResponseWithBody(request, http.StatusOK, response)
	})

	var rawResponse *http.Response
	ctx := runtime.WithCaptureResponse(*mockContext.Context, &rawResponse)

	azCli := newAzCliFromMockContext(mockContext)
	// We don't care about the actual response or if an error occurred
	// Any API call that leverages the Go SDK is fine
	_, _ = azCli.GetResource(ctx, "SUBSCRIPTION_ID", "RESOURCE_ID", "API_VERSION")

	userAgent, ok := rawResponse.Request.Header["User-Agent"]
	if !ok {
		require.Fail(t, "missing User-Agent header")
	}

	require.Contains(t, userAgent[0], "azsdk-go")
	require.Contains(t, userAgent[0], "azdev")
}

// NewAzCliFromMockContext creates a new instance of AzCli, configured to use the credential and pipeline from the
// provided mock context.
// TODO: this is duplicated in mocks.go... what refactor can be had here?
func newAzCliFromMockContext(mockContext *mocks.MockContext) AzCli {
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

	return NewAzCli(
		mockaccount.SubscriptionCredentialProviderFunc(func(_ context.Context, _ string) (azcore.TokenCredential, error) {
			return mockContext.Credentials, nil
		}),
		mockContext.HttpClient,
		NewAzCliArgs{},
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
