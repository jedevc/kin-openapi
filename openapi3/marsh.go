package openapi3

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/oasdiff/yaml"
)

func unmarshalError(jsonUnmarshalErr error) error {
	if before, after, found := strings.Cut(jsonUnmarshalErr.Error(), "Bis"); found && before != "" && after != "" {
		before = strings.ReplaceAll(before, " Go struct ", " ")
		return fmt.Errorf("%s%s", before, strings.ReplaceAll(after, "Bis", ""))
	}
	return jsonUnmarshalErr
}

// unmarshal decodes data into v. It returns the document origin tree when
// includeOrigin is set and the data took the yaml path (json input carries no
// origins), so the caller can retain it (see Loader.originTrees).
func unmarshal(data []byte, v any, includeOrigin bool, location *url.URL) (*originTree, error) {
	var jsonErr, yamlErr error

	// See https://github.com/getkin/kin-openapi/issues/680
	if jsonErr = json.Unmarshal(data, v); jsonErr == nil {
		return nil, nil
	}

	// UnmarshalStrict(data, v) TODO: investigate how ymlv3 handles duplicate map keys
	var file string
	if location != nil {
		file = location.String()
	}
	// Origin tracking is always requested here, regardless of includeOrigin:
	// the position data it carries is also how applyOrigins restores true
	// document order on every ordered-map-backed collection, which the YAML
	// decode above loses along the way (see convertToJSONableObject in
	// oasdiff/yaml, which round-trips through a plain, unordered Go map).
	// setOrigin gates only whether Origin fields actually get set on the
	// document, keeping that part of the contract unchanged for callers that
	// didn't ask for it.
	if tree, err := yaml.Unmarshal(data, v, yaml.DecodeOpts{
		Origin:            yaml.OriginOpt{Enabled: true, File: file},
		DisableTimestamps: true,
	}); err == nil {
		applyOrigins(v, tree, includeOrigin)
		if !includeOrigin {
			return nil, nil
		}
		return tree, nil
	} else {
		yamlErr = err
	}

	// If both unmarshaling attempts fail, return a new error that includes both errors
	return nil, fmt.Errorf("failed to unmarshal data: json error: %v, yaml error: %v", jsonErr, yamlErr)
}
