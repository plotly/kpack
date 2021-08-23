package cosigner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/lifecycle"
	"github.com/sigstore/cosign/cmd/cosign/cli"
)

type ImageSigner struct {
	Logger *log.Logger
}

var (
	cliSignCmd     = cli.SignCmd
	secretLocation = "/var/build-secrets"
)

// Other keyops support: https://github.com/sigstore/cosign/blob/143e47a120702f175e68e0a04594cb87a4ce8e02/cmd/cosign/cli/sign.go#L167
// Todo: Annotation obtained from kpack config

func NewImageSigner(logger *log.Logger) *ImageSigner {
	return &ImageSigner{
		Logger: logger,
	}
}

// signCmd will just use the mounted file instead of trying to access kuberenets for the secret
func (s *ImageSigner) Sign(reportFilePath string, annotations map[string]interface{}) error {
	var report lifecycle.ExportReport
	_, err := toml.DecodeFile(reportFilePath, &report)
	if err != nil {
		return fmt.Errorf("toml decode: %v", err)
	}

	if len(report.Image.Tags) < 1 {
		s.Logger.Println("no image tag to sign")
		return nil
	}

	cosignFiles := findCosignFiles(secretLocation, s.Logger)
	if len(cosignFiles) == 0 {
		s.Logger.Println("no keys found for cosign signing")
		return nil
	}

	refImage := report.Image.Tags[0]

	ctx := context.Background()
	for _, cosignFile := range cosignFiles {
		ko := cli.KeyOpts{KeyRef: cosignFile, PassFunc: func(bool) ([]byte, error) {
			// Todo: Use cosign.password
			return []byte(""), nil
		}}

		if err := cliSignCmd(ctx, ko, annotations, refImage, "", true, "", false, false); err != nil {
			return fmt.Errorf("unable to sign image with %s: %v", cosignFile, err)
		}
	}

	return nil
}

func findCosignFiles(dir string, logger *log.Logger) []string {
	var files []string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil || f == nil {
			return nil
		}

		if !f.IsDir() {
			// Only look at /secretLocation/folder/cosign.key folder/file structure
			r, err := regexp.MatchString(`^`+regexp.QuoteMeta(secretLocation)+`\/[^\/]+\/cosign.key`, path)
			if err == nil && r {
				files = append(files, path)
			}
		}
		return nil
	})
	return files
}
