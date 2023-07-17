package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hashicorp/go-tfe"
)

// GetPrivateProvider returns the provider if it exists, otherwise nil.
func GetPrivateProvider(c *tfe.Client, id tfe.RegistryProviderID) (*tfe.RegistryProvider, error) {
	resp, err := c.RegistryProviders.Read(
		context.Background(),
		id,
		&tfe.RegistryProviderReadOptions{},
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreatePrivateProvider creates a private provider in the organization.
func CreatePrivateProvider(c *tfe.Client, id tfe.RegistryProviderID) (*tfe.RegistryProvider, error) {
	p, err := c.RegistryProviders.Create(
		context.Background(),
		id.OrganizationName,
		tfe.RegistryProviderCreateOptions{
			Name:         id.Name,
			Namespace:    id.Namespace,
			RegistryName: id.RegistryName,
		},
	)
	if err != nil {
		return nil, err
	}

	return p, err
}

// GetPrivateProviderVersion returns the version if it exists, otherwise nil.
func GetPrivateProviderVersion(c *tfe.Client, version tfe.RegistryProviderVersionID) (*tfe.RegistryProviderVersion, error) {
	resp, err := c.RegistryProviderVersions.Read(
		context.Background(),
		version,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreatePrivateProviderVersion creates a version of the private provider.
func CreatePrivateProviderVersion(c *tfe.Client, version tfe.RegistryProviderVersionID, gpgKeyID string) (*tfe.RegistryProviderVersion, error) {
	v, err := c.RegistryProviderVersions.Create(
		context.Background(),
		version.RegistryProviderID,
		tfe.RegistryProviderVersionCreateOptions{
			Version: version.Version,
			KeyID:   gpgKeyID,
		},
	)
	if err != nil {
		return nil, err
	}

	return v, err
}

func GetPrivateProviderPlatform(c *tfe.Client, plat tfe.RegistryProviderPlatformID) (*tfe.RegistryProviderPlatform, error) {
	resp, err := c.RegistryProviderPlatforms.Read(
		context.Background(),
		plat,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func CreatePrivateProviderPlatform(c *tfe.Client, plat tfe.RegistryProviderPlatformID, shasum, filename string) (*tfe.RegistryProviderPlatform, error) {
	resp, err := c.RegistryProviderPlatforms.Create(
		context.Background(),
		plat.RegistryProviderVersionID,
		tfe.RegistryProviderPlatformCreateOptions{
			OS:       plat.OS,
			Arch:     plat.Arch,
			Shasum:   shasum,
			Filename: filename,
		},
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func UploadFile(ctx context.Context, url, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, f)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("received %d instead of 200: %s", resp.StatusCode, string(body))
	}

	return nil
}
