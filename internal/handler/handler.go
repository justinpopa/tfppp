package handler

import "github.com/hashicorp/go-tfe"

type Handler struct {
	Client                  *tfe.Client
	GPG_KeyID               string
	RegistryProviderVersion tfe.RegistryProviderVersionID
	Metadata                *Metadata
	Artifacts               *[]Artifact
}

func NewHandler(token, gpg_keyid, org, name, ver string) (*Handler, error) {
	config := tfe.DefaultConfig()
	config.Token = token

	c, err := tfe.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Handler{
		Client: c,
		RegistryProviderVersion: tfe.RegistryProviderVersionID{
			RegistryProviderID: tfe.RegistryProviderID{
				OrganizationName: org,
				RegistryName:     tfe.PrivateRegistry,
				Namespace:        org,
				Name:             name,
			},
			Version: ver,
		},
		GPG_KeyID: gpg_keyid,
	}, nil
}
