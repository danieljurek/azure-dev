package azcli

import (
	"context"
	"fmt"
	"io"

	"github.com/azure/azure-dev/cli/azd/pkg/convert"
)

type AzCliFunctionAppProperties struct {
	HostNames []string
}

func (cli *azCli) GetFunctionAppProperties(
	ctx context.Context,
	subscriptionId string,
	resourceGroup string,
	appName string,
) (*AzCliFunctionAppProperties, error) {
	webApp, err := cli.webAppsClient.Get(ctx, resourceGroup, appName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving function app properties: %w", err)
	}

	return &AzCliFunctionAppProperties{
		HostNames: []string{*webApp.Properties.DefaultHostName},
	}, nil
}

func (cli *azCli) DeployFunctionAppUsingZipFile(
	ctx context.Context,
	subscriptionId string,
	resourceGroup string,
	appName string,
	deployZipFile io.Reader,
) (*string, error) {
	response, err := cli.zipDeployClient.Deploy(ctx, appName, deployZipFile)
	if err != nil {
		return nil, err
	}

	return convert.RefOf(response.StatusText), nil
}
