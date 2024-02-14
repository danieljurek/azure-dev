package azcli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/azure/azure-dev/cli/azd/pkg/convert"
)

type AzCliKeyVault struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	Properties struct {
		EnableSoftDelete      bool `json:"enableSoftDelete"`
		EnablePurgeProtection bool `json:"enablePurgeProtection"`
	} `json:"properties"`
}

type AzCliKeyVaultSecret struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (cli *azCli) GetKeyVault(
	ctx context.Context,
	resourceGroupName string,
	vaultName string,
) (*AzCliKeyVault, error) {
	vault, err := cli.keyVaultsClient.Get(ctx, resourceGroupName, vaultName, nil)
	if err != nil {
		return nil, fmt.Errorf("getting key vault: %w", err)
	}

	return &AzCliKeyVault{
		Id:       *vault.ID,
		Name:     *vault.Name,
		Location: *vault.Location,
		Properties: struct {
			EnableSoftDelete      bool "json:\"enableSoftDelete\""
			EnablePurgeProtection bool "json:\"enablePurgeProtection\""
		}{
			EnableSoftDelete:      convert.ToValueWithDefault(vault.Properties.EnableSoftDelete, false),
			EnablePurgeProtection: convert.ToValueWithDefault(vault.Properties.EnablePurgeProtection, false),
		},
	}, nil
}

func (cli *azCli) GetKeyVaultSecret(
	ctx context.Context,
	vaultName string,
	secretName string,
) (*AzCliKeyVaultSecret, error) {
	vaultUrl := vaultName
	if !strings.Contains(strings.ToLower(vaultName), "https://") {
		vaultUrl = fmt.Sprintf("https://%s.vault.azure.net", vaultName)
	}

	client, err := cli.secretsClientFactory(vaultUrl)
	if err != nil {
		return nil, err
	}
	response, err := client.GetSecret(ctx, secretName, "", nil)
	if err != nil {
		var httpErr *azcore.ResponseError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, ErrAzCliSecretNotFound
		}
		return nil, fmt.Errorf("getting key vault secret: %w", err)
	}

	return &AzCliKeyVaultSecret{
		Id:    response.SecretBundle.ID.Version(),
		Name:  response.SecretBundle.ID.Name(),
		Value: *response.SecretBundle.Value,
	}, nil
}

func (cli *azCli) PurgeKeyVault(ctx context.Context, vaultName string, location string) error {
	poller, err := cli.keyVaultsClient.BeginPurgeDeleted(ctx, vaultName, location, nil)
	if err != nil {
		return fmt.Errorf("starting purging key vault: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("purging key vault: %w", err)
	}

	return nil
}
