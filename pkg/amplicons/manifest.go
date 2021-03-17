package amplicons

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

var (

	// ErrNoSchemeName is returned when primer scheme name is not provided
	ErrNoSchemeName = errors.New("no primer scheme name provided")

	// ErrNoSchemeVersion is returned when a bad or non-existant scheme version is provided
	ErrNoSchemeVersion = errors.New("request scheme version must be >= 0")
)

// GetManifest will download the ARTIC primer scheme
// manifest and populate an in memory manifest.
func GetManifest(manifestURL string) (*api.Manifest, error) {

	// create an empty manifest to populate
	manifest := &api.Manifest{
		Schemes: make(map[string]*api.SchemeMetadata),
	}

	// download the json
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode the json into the manifest struct
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

// CheckManifest will check a scheme and version are found in the manifest.
// It will return the scheme tag to use to download this scheme, or any
// error in checking the manifest.
func CheckManifest(manifest *api.Manifest, requestedScheme string, requestedVersion int32) (string, error) {
	if len(requestedScheme) == 0 {
		return "", ErrNoSchemeName
	}
	if requestedVersion < 0 {
		return "", ErrNoSchemeVersion
	}
	schemeTag := ""
loop:
	for name, schemeMetadata := range manifest.GetSchemes() {
		for _, alias := range schemeMetadata.GetAliases() {
			if requestedScheme == alias {
				if requestedVersion <= schemeMetadata.GetLatestVersion() {
					schemeTag = name
					break loop
				}
			}
		}
	}
	if schemeTag == "" {
		return "", fmt.Errorf("can't find scheme in manifest for %v, version %d", requestedScheme, requestedVersion)
	}
	return schemeTag, nil
}
