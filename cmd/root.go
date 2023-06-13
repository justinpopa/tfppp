package cmd

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/justinpopa/goreleaser-tfpp/internal/handler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const GORELEASER_MANIFEST_FILE = "dist/metadata.json"
const GORELEASER_ARTIFACTS_FILE = "dist/artifacts.json"

var rootCmd = &cobra.Command{
	Use:   "goreleaser-tfpp",
	Short: "Custom publisher for goreleaser to upload the necessary files to Terraform Cloud",
	Run: func(cmd *cobra.Command, args []string) {
		h, err := handler.NewHandler(
			viper.GetString("TFC_TOKEN"),
			viper.GetString("GPG_KEY_ID"),
			viper.GetString("TFC_ORG"),
			viper.GetString("PROVIDER_NAME"),
			viper.GetString("PROVIDER_VERSION"),
		)
		if err != nil {
			panic(err)
		}

		// check if manifest file exists and load it
		_, err = os.Stat(GORELEASER_MANIFEST_FILE)
		if os.IsNotExist(err) {
			panic(err)
		} else {
			err = h.GetMetadata(GORELEASER_MANIFEST_FILE)
			if err != nil {
				panic(err)
			}
		}

		// check if artifacts file exists and load it
		_, err = os.Stat(GORELEASER_ARTIFACTS_FILE)
		if os.IsNotExist(err) {
			panic(err)
		} else {
			err = h.GetArtifacts(GORELEASER_ARTIFACTS_FILE)
			if err != nil {
				panic(err)
			}
		}

		// find or create private provider
		provider, err := h.GetPrivateProvider()
		if err != nil && err.Error() != "resource not found" {
			panic(err)
		}

		if provider == nil {
			log.Printf("Creating provider %s\n", h.Metadata.Project)
			provider, err = h.CreatePrivateProvider()
			if err != nil {
				panic(err)
			}
		}

		// find or create private provider version
		version, err := h.GetPrivateProviderVersion()
		if err != nil && err.Error() != "resource not found" {
			panic(err)
		}

		if version == nil {
			log.Printf("Creating version %s\n", h.Metadata.Tag)
			version, err = h.CreatePrivateProviderVersion()
			if err != nil {
				panic(err)
			}
		} else {
			log.Printf("Version %s already exists\n", h.Metadata.Tag)
		}

		// upload shasum and signature files
		if !version.ShasumsUploaded {
			log.Println("Uploading checksums")

			// get the url to upload the checksum file to
			sums_upload_url, err := version.ShasumsUploadURL()
			if err != nil {
				panic(err)
			}

			// upload the checksum file
			err = handler.UploadFile(context.Background(), sums_upload_url, *h.GetShaSums())
			if err != nil {
				panic(err)
			}
		} else {
			log.Println("Checksums already exist")
		}

		if !version.ShasumsSigUploaded {
			log.Println("Uploading signatures")

			// get the url to upload the checksum file to
			sigs_upload_url, err := version.ShasumsSigUploadURL()
			if err != nil {
				panic(err)
			}

			// upload the checksum file
			err = handler.UploadFile(context.Background(), sigs_upload_url, *h.GetSumsSig())
			if err != nil {
				panic(err)
			}
		} else {
			log.Println("Signatures already exist")
		}

		// loop through the archive artifacts and upload them
		for _, a := range *h.Artifacts {
			if a.Type == "Archive" {
				// TODO: check if the platform exists already, and if not, create it
				platform, err := h.Client.RegistryProviderPlatforms.Read(
					context.Background(),
					tfe.RegistryProviderPlatformID{
						RegistryProviderVersionID: h.RegistryProviderVersion,
						OS:                        a.GoOS,
						Arch:                      a.GoArch,
					},
				)
				if err != nil && err.Error() != "resource not found" {
					panic(err)
				}

				if platform != nil {
					log.Printf("%s_%s already exists\n", a.GoOS, a.GoArch)
				} else {
					log.Printf("Creating platform: %s_%s\n", a.GoOS, a.GoArch)

					platform, err = h.Client.RegistryProviderPlatforms.Create(
						context.Background(),
						h.RegistryProviderVersion,
						tfe.RegistryProviderPlatformCreateOptions{
							OS:       a.GoOS,
							Arch:     a.GoArch,
							Shasum:   strings.Split(a.Extra.Checksum, ":")[1],
							Filename: a.Name,
						},
					)
					if err != nil {
						panic(err)
					}
				}

				if !platform.ProviderBinaryUploaded {
					log.Printf("Uploading %s\n", a.Name)

					// upload the artifact
					err = handler.UploadFile(
						context.Background(),
						platform.Links["provider-binary-upload"].(string),
						a.Path,
					)
					if err != nil {
						panic(err)
					}

				} else {
					log.Printf("%s already uploaded, skipping.\n", a.Name)
				}
			}
		}
	},
}

func Execute() error {
	err := rootCmd.Execute()
	return err
}

func init() {
	cobra.OnInitialize(initEnvs)

	flags := rootCmd.Flags()

	flags.String("tfc_token", "", "Terraform Cloud API token used to publish the provider")
	viper.BindPFlag("TFC_TOKEN", flags.Lookup("tfc_token"))

	flags.String("tfc_org", "", "Terraform Cloud organization the provider is published to")
	viper.BindPFlag("TFC_ORG", flags.Lookup("tfc_org"))

	flags.String("gpg_key_id", "", "GPG key ID used to sign the provider binaries")
	viper.BindPFlag("GPG_KEY_ID", flags.Lookup("gpg_key_id"))

	flags.String("provider_name", "", "Name of the provider")
	viper.BindPFlag("PROVIDER_NAME", flags.Lookup("provider_name"))

	flags.String("project_version", "", "Version of the provider")
	viper.BindPFlag("PROVIDER_VERSION", flags.Lookup("provider_version"))
}

func initEnvs() {
	viper.AutomaticEnv()
}
