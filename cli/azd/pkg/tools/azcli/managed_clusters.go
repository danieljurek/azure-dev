package azcli

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
)

// ManagedClustersService provides actions on top of Azure Kubernetes Service (AKS) Managed Clusters
type ManagedClustersService interface {
	// Gets the admin credentials for the specified resource
	GetAdminCredentials(
		ctx context.Context,
		subscriptionId string,
		resourceGroupName string,
		resourceName string,
	) (*armcontainerservice.CredentialResults, error)
	// Gets the user credentials for the specified resource
	GetUserCredentials(
		ctx context.Context,
		subscriptionId string,
		resourceGroupName string,
		resourceName string,
	) (*armcontainerservice.CredentialResults, error)
}

type managedClustersService struct {
	managedClustersClient *armcontainerservice.ManagedClustersClient
}

// Creates a new instance of the ManagedClustersService
func NewManagedClustersService(
	managedClustersClient *armcontainerservice.ManagedClustersClient,
) ManagedClustersService {
	return &managedClustersService{
		managedClustersClient: managedClustersClient,
	}
}

// Gets the user credentials for the specified resource
func (cs *managedClustersService) GetUserCredentials(
	ctx context.Context,
	subscriptionId string,
	resourceGroupName string,
	resourceName string,
) (*armcontainerservice.CredentialResults, error) {
	credResult, err := cs.managedClustersClient.ListClusterUserCredentials(ctx, resourceGroupName, resourceName, nil)
	if err != nil {
		return nil, err
	}

	return &credResult.CredentialResults, nil
}

// Gets the admin credentials for the specified resource
func (cs *managedClustersService) GetAdminCredentials(
	ctx context.Context,
	subscriptionId string,
	resourceGroupName string,
	resourceName string,
) (*armcontainerservice.CredentialResults, error) {
	credResult, err := cs.managedClustersClient.ListClusterAdminCredentials(ctx, resourceGroupName, resourceName, nil)
	if err != nil {
		return nil, err
	}

	return &credResult.CredentialResults, nil
}
