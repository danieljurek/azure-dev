package azcli

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appplatform/armappplatform/v2"
	"github.com/Azure/azure-storage-file-go/azfile"
)

// SpringService provides artifacts upload/deploy and query to Azure Spring Apps (ASA)
type SpringService interface {
	// Get Spring app properties
	GetSpringAppProperties(
		ctx context.Context,
		resourceGroupName string,
		instanceName string,
		appName string,
	) (*SpringAppProperties, error)
	// Deploy jar artifact to ASA app deployment
	DeploySpringAppArtifact(
		ctx context.Context,
		resourceGroup string,
		instanceName string,
		appName string,
		relativePath string,
		deploymentName string,
	) (*string, error)
	// Upload jar artifact to ASA app Storage File
	UploadSpringArtifact(
		ctx context.Context,
		resourceGroup string,
		instanceName string,
		appName string,
		artifactPath string,
	) (*string, error)
	// Get Spring app deployment
	GetSpringAppDeployment(
		ctx context.Context,
		resourceGroupName string,
		instanceName string,
		appName string,
		deploymentName string,
	) (*string, error)
}

type springService struct {
	appsClient        *armappplatform.AppsClient
	deploymentsClient *armappplatform.DeploymentsClient
}

// Creates a new instance of the NewSpringService
func NewSpringService(
	appsClient *armappplatform.AppsClient,
	deploymentsClient *armappplatform.DeploymentsClient,
) SpringService {
	return &springService{
		appsClient:        appsClient,
		deploymentsClient: deploymentsClient,
	}
}

type SpringAppProperties struct {
	Url []string
}

func (ss *springService) GetSpringAppProperties(
	ctx context.Context,
	resourceGroup, instanceName, appName string,
) (*SpringAppProperties, error) {
	springApp, err := ss.appsClient.Get(ctx, resourceGroup, instanceName, appName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving spring app properties: %w", err)
	}

	var url []string
	if springApp.Properties != nil &&
		springApp.Properties.URL != nil &&
		*springApp.Properties.Public {
		url = []string{*springApp.Properties.URL}
	} else {
		url = []string{}
	}

	return &SpringAppProperties{
		Url: url,
	}, nil
}

func (ss *springService) UploadSpringArtifact(
	ctx context.Context,
	resourceGroup, instanceName, appName, artifactPath string,
) (*string, error) {
	file, err := os.Open(artifactPath)

	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("artifact %s does not exist: %w", artifactPath, err)
	}
	if err != nil {
		return nil, fmt.Errorf("reading artifact file %s: %w", artifactPath, err)
	}
	defer file.Close()

	storageInfo, err := ss.appsClient.GetResourceUploadURL(ctx, resourceGroup, instanceName, appName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource upload URL: %w", err)
	}

	url, err := url.Parse(*storageInfo.UploadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse storage upload url %s : %w", *storageInfo.UploadURL, err)
	}

	// Pass NewAnonymousCredential here, since the URL returned by Azure Spring Apps already contains a SAS token
	fileURL := azfile.NewFileURL(*url, azfile.NewPipeline(azfile.NewAnonymousCredential(), azfile.PipelineOptions{}))
	err = azfile.UploadFileToAzureFile(ctx, file, fileURL,
		azfile.UploadToAzureFileOptions{
			Metadata: azfile.Metadata{
				"createdby": "AZD",
			},
		})

	if err != nil {
		return nil, fmt.Errorf("failed to upload artifact %s : %w", artifactPath, err)
	}

	return storageInfo.RelativePath, nil
}

func (ss *springService) DeploySpringAppArtifact(
	ctx context.Context,
	resourceGroup string,
	instanceName string,
	appName string,
	relativePath string,
	deploymentName string,
) (*string, error) {
	_, err := ss.createOrUpdateDeployment(
		ctx,
		resourceGroup,
		instanceName,
		appName,
		deploymentName,
		relativePath,
	)
	if err != nil {
		return nil, err
	}
	resName, err := ss.activeDeployment(ctx, resourceGroup, instanceName, appName, deploymentName)
	if err != nil {
		return nil, err
	}

	return resName, nil
}

func (ss *springService) GetSpringAppDeployment(
	ctx context.Context,
	resourceGroupName string,
	instanceName string,
	appName string,
	deploymentName string,
) (*string, error) {
	resp, err := ss.deploymentsClient.Get(ctx, resourceGroupName, instanceName, appName, deploymentName, nil)

	if err != nil {
		return nil, err
	}

	return resp.Name, nil
}

func (ss *springService) createOrUpdateDeployment(
	ctx context.Context,
	resourceGroup string,
	instanceName string,
	appName string,
	deploymentName string,
	relativePath string,
) (*string, error) {
	poller, err := ss.deploymentsClient.BeginCreateOrUpdate(ctx, resourceGroup, instanceName, appName, deploymentName,
		armappplatform.DeploymentResource{
			Properties: &armappplatform.DeploymentResourceProperties{
				Source: &armappplatform.JarUploadedUserSourceInfo{
					Type:         to.Ptr("Jar"),
					RelativePath: to.Ptr(relativePath),
				},
			},
		}, nil)
	if err != nil {
		return nil, err
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return res.Name, nil
}

func (ss *springService) activeDeployment(
	ctx context.Context,
	resourceGroup string,
	instanceName string,
	appName string,
	deploymentName string,
) (*string, error) {
	poller, err := ss.appsClient.BeginSetActiveDeployments(ctx, resourceGroup, instanceName, appName,
		armappplatform.ActiveDeploymentCollection{
			ActiveDeploymentNames: []*string{
				to.Ptr(deploymentName),
			},
		}, nil)

	if err != nil {
		return nil, err
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return res.Name, nil
}
