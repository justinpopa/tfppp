package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hashicorp/go-tfe"
)

const RegistryName = tfe.PrivateRegistry
const PageSize = 25

// GetPrivateProviders returns a list of all private providers in the organization.
func (h *Handler) GetPrivateProviders() ([]tfe.RegistryProvider, error) {
	providers_options := tfe.RegistryProviderListOptions{
		RegistryName: RegistryName,
		ListOptions: tfe.ListOptions{
			PageNumber: 1,
			PageSize:   PageSize,
		},
	}

	providers := []tfe.RegistryProvider{}

	for {
		resp, err := h.Client.RegistryProviders.List(
			context.Background(),
			h.RegistryProviderVersion.RegistryProviderID.OrganizationName,
			&providers_options,
		)
		if err != nil {
			return nil, err
		}

		for _, provider := range resp.Items {
			providers = append(providers, *provider)
		}

		if resp.Pagination.CurrentPage >= resp.Pagination.TotalPages {
			break
		}

		providers_options.PageNumber = resp.Pagination.NextPage
	}

	return providers, nil
}

// GetPrivateProvider returns the provider if it exists, otherwise nil.
func (h *Handler) GetPrivateProvider() (*tfe.RegistryProvider, error) {
	resp, err := h.Client.RegistryProviders.Read(
		context.Background(),
		h.RegistryProviderVersion.RegistryProviderID,
		&tfe.RegistryProviderReadOptions{},
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreatePrivateProvider creates a private provider in the organization.
func (h *Handler) CreatePrivateProvider() (*tfe.RegistryProvider, error) {
	p, err := h.Client.RegistryProviders.Create(
		context.Background(),
		h.RegistryProviderVersion.RegistryProviderID.OrganizationName,
		tfe.RegistryProviderCreateOptions{
			Name:         h.RegistryProviderVersion.RegistryProviderID.Name,
			Namespace:    h.RegistryProviderVersion.RegistryProviderID.Namespace,
			RegistryName: RegistryName,
		},
	)
	if err != nil {
		return nil, err
	}

	return p, err
}

// GetPrivateProviderVersions returns a list of all versions of the private provider.
func (h *Handler) GetPrivateProviderVersions() ([]tfe.RegistryProviderVersion, error) {
	versions_options := tfe.RegistryProviderVersionListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 1,
			PageSize:   PageSize,
		},
	}

	versions := []tfe.RegistryProviderVersion{}

	for {
		resp, err := h.Client.RegistryProviderVersions.List(
			context.Background(),
			h.RegistryProviderVersion.RegistryProviderID,
			&versions_options,
		)
		if err != nil {
			return nil, err
		}

		for _, version := range resp.Items {
			versions = append(versions, *version)
		}

		if resp.Pagination.CurrentPage >= resp.Pagination.TotalPages {
			break
		}

		versions_options.PageNumber = resp.Pagination.NextPage
	}

	return versions, nil
}

// GetPrivateProviderVersion returns the version if it exists, otherwise nil.
func (h *Handler) GetPrivateProviderVersion() (*tfe.RegistryProviderVersion, error) {
	resp, err := h.Client.RegistryProviderVersions.Read(
		context.Background(),
		h.RegistryProviderVersion,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreatePrivateProviderVersion creates a version of the private provider.
func (h *Handler) CreatePrivateProviderVersion() (*tfe.RegistryProviderVersion, error) {
	v, err := h.Client.RegistryProviderVersions.Create(
		context.Background(),
		h.RegistryProviderVersion.RegistryProviderID,
		tfe.RegistryProviderVersionCreateOptions{
			Version: h.RegistryProviderVersion.Version,
			KeyID:   h.GPG_KeyID,
		},
	)
	if err != nil {
		return nil, err
	}

	return v, err
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
