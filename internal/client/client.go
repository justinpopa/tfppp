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
func GetPrivateProvider(id tfe.RegistryProviderID) (*tfe.RegistryProvider, error) {
	c := tfe.Client{}
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
func CreatePrivateProvider(id tfe.RegistryProviderID) (*tfe.RegistryProvider, error) {
	c := tfe.Client{}
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
func GetPrivateProviderVersion(version tfe.RegistryProviderVersionID) (*tfe.RegistryProviderVersion, error) {
	c := tfe.Client{}
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
func CreatePrivateProviderVersion(version tfe.RegistryProviderVersionID, gpgKeyID string) (*tfe.RegistryProviderVersion, error) {
	c := tfe.Client{}
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

func GetPrivateProviderPlatform(plat tfe.RegistryProviderPlatformID) (*tfe.RegistryProviderPlatform, error) {
	c := tfe.Client{}
	resp, err := c.RegistryProviderPlatforms.Read(
		context.Background(),
		plat,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func CreatePrivateProviderPlatform(plat tfe.RegistryProviderPlatformID, shasum, filename string) (*tfe.RegistryProviderPlatform, error) {
	c := tfe.Client{}
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
