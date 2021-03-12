// Package amplicons performs the amplicon scheme validation and filtering for Archer.
package amplicons

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

var (
	ErrNoSchemeName    = errors.New("no primer scheme name provided")
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
func CheckManifest(manifest *api.Manifest, requestedScheme string, requestedVersion int32) error {
	if len(requestedScheme) == 0 {
		return ErrNoSchemeName
	}
	if requestedVersion < 0 {
		return ErrNoSchemeVersion
	}
	ok := false
loop:
	for _, schemeMetadata := range manifest.GetSchemes() {
		for _, alias := range schemeMetadata.GetAliases() {
			if requestedScheme == alias {
				if requestedVersion <= schemeMetadata.GetLatestVersion() {
					ok = true
					break loop
				}
			}
		}
	}
	if !ok {
		return fmt.Errorf("can't find scheme in manifest for %v, version %d", requestedScheme, requestedVersion)
	}
	return nil
}
