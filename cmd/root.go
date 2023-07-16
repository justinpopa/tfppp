package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/justinpopa/tfppp/internal/artifacts"
	"github.com/justinpopa/tfppp/internal/client"
	"github.com/spf13/cobra"
)

const GORELEASER_ARTIFACTS_FILE = "artifacts.json"

var rootCmd = &cobra.Command{
	Use:   "tfppp",
	Short: "Terraform Private Provider Publisher",
	Long:  `Custom publisher for goreleaser that takes the dist folder output from goreleaser and a few env vars to publish private providers to TFC.`,
	Args:  cobra.ExactArgs(6),
	Run: func(cmd *cobra.Command, args []string) {

		/* Set up Provider */
		org, err := cmd.Flags().GetString("organization")
		if err != nil {
			panic(err) // TODO: handle error
		}

		fullName, err := cmd.Flags().GetString("name")
		if err != nil {
			panic(err) // TODO: handle error
		}

		name := strings.Replace(fullName, "terraform-provider-", "", 1)

		providerID := tfe.RegistryProviderID{
			OrganizationName: org,
			RegistryName:     tfe.PrivateRegistry,
			Namespace:        org,
			Name:             name,
		}

		// find or create private provider
		provider, err := client.GetPrivateProvider(
			providerID,
		)
		if err != nil && err.Error() != "resource not found" {
			panic(err) // TODO: handle error
		}

		if provider == nil {
			log.Printf("Provider %s not found. Creating.\n", name)
			provider, err = client.CreatePrivateProvider(
				providerID,
			)
			if err != nil {
				panic(err) // TODO: handle error
			}
		}

		/* Set up Provider Version */

		ver, err := cmd.Flags().GetString("version")
		if err != nil {
			panic(err) // TODO: handle error
		}

		fingerprint, err := cmd.Flags().GetString("fingerprint")
		if err != nil {
			panic(err) // TODO: handle error
		}

		artifactSlice, err := artifacts.Get(GORELEASER_ARTIFACTS_FILE)
		if err != nil {
			panic(err) // TODO: handle error
		}

		versionId := tfe.RegistryProviderVersionID{
			RegistryProviderID: tfe.RegistryProviderID{
				OrganizationName: provider.Organization.Name,
				RegistryName:     tfe.PrivateRegistry,
				Namespace:        provider.Namespace,
				Name:             provider.Name,
			},
			Version: ver,
		}

		// find private provider version, if it exists. if not, create it.
		version, err := client.GetPrivateProviderVersion(versionId)
		if err != nil && err.Error() != "resource not found" {
			panic(err) // TODO: handle error
		}

		if version == nil {
			log.Printf("Version %v/%v not found. Creating.\n", name, ver)
			version, err = client.CreatePrivateProviderVersion(versionId, fingerprint)
			if err != nil {
				panic(err) // TODO: handle error
			}
		}

		/* Sums and Sigs Upload */
		if !version.ShasumsUploaded {
			log.Println("Shasums have not been uploaded, uploading now.")

			// get the url to upload the checksum file to
			sumsUploadURL, err := version.ShasumsUploadURL()
			if err != nil {
				panic(err) // TODO: handle error
			}

			// get the checksum file out of the artifactSlice
			for _, a := range *artifactSlice {
				if a.Type == "Checksum" {
					err = client.UploadFile(context.Background(), sumsUploadURL, a.Path)
					if err != nil {
						panic(err) // TODO: handle error
					}
				}
			}
		}

		if !version.ShasumsSigUploaded {
			log.Println("Shasums signatures have not been uploaded, uploading now.")

			// get the url to upload the signatures file to
			sigsUploadURL, err := version.ShasumsSigUploadURL()
			if err != nil {
				panic(err) // TODO: handle error
			}

			// get the signatures file out of the artifactSlice
			for _, a := range *artifactSlice {
				if a.Type == "Checksum" {
					err = client.UploadFile(context.Background(), sigsUploadURL, a.Path)
					if err != nil {
						panic(err) // TODO: handle error
					}
				}
			}
		}

		/* Platform creation and upload */

		filename, err := cmd.Flags().GetString("artifact")
		if err != nil {
			panic(err) // TODO: handle error
		}

		var artifact artifacts.Artifact

		// get the platform from the artifactSlice
		for _, a := range *artifactSlice {
			if a.Type == "Archive" {
				if strings.Contains(a.Path, filename) {
					artifact = a
				}
			}
		}

		platformID := tfe.RegistryProviderPlatformID{
			RegistryProviderVersionID: versionId,
			OS:                        artifact.GoOS,
			Arch:                      artifact.GoArch,
		}

		// find platform if it exists. if not, create it.
		platform, err := client.GetPrivateProviderPlatform(platformID)
		if err != nil && err.Error() != "resource not found" {
			panic(err) // TODO: handle error
		}

		if platform == nil {
			log.Printf("Platform %v/%v/%v_%v not found. Creating.\n", name, ver, platformID.OS, platformID.Arch)
			platform, err = client.CreatePrivateProviderPlatform(platformID, artifact.Extra.Checksum, filename)
			if err != nil {
				panic(err) // TODO: handle error
			}
		}

		// upload the artifact if it hasn't been uploaded yet
		if !platform.ProviderBinaryUploaded {
			log.Printf("Uploading %s\n", filename)

			// upload the artifact
			err := client.UploadFile(
				context.Background(),
				platform.Links["provider-binary-upload"].(string),
				filename,
			)
			if err != nil {
				panic(err) // TODO: handle error
			}
		}
	},
}

func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add flags
	rootCmd.Flags().StringP("artifact", "a", "", "Artifact name, e.g.: terraform-provider-hashicups_0.0.1_darwin_arm64.zip")
	rootCmd.Flags().StringP("name", "n", "", "Full project name, e.g.: terraform-provider-hashicups")
	rootCmd.Flags().StringP("version", "v", "", "Version, e.g.: 0.1.1")
	rootCmd.Flags().StringP("fingerprint", "f", "", "GPG fingerprint")
	rootCmd.Flags().StringP("organization", "o", "", "TFC Organization")
	rootCmd.Flags().StringP("token", "t", "", "TFC Token")
}
