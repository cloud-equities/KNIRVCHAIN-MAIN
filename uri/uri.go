package uri

import (
	"errors"
	//"chain"
	//"fmt"
	//"knirv"
	"strings"
)

// ParseURI parses a KNIRV URI
func ParseURI(uri string) (string, string, error) {
	if !strings.HasPrefix(uri, "knirv://") {
		return "", "", errors.InvalidURIError
	}

	parts := strings.Split(uri[8:], ".")

	if len(parts) != 2 {
		return "", "", errors.InvalidURIError
	}

	resource := parts[1]
	identifier := parts[0]
	return resource, identifier, nil
}

// ResolveURI resolves a KNIRV URI, return the corresponding object
//func ResolveURI(uri string, chain *chain.Block) (interface{}, error) {
//resource, identifier, err := ParseURI(uri)
//if err != nil {
//	return nil, fmt.Errorf("failed to parse uri: %w", err)
//	}

//switch resource {
//case "nrn":
//nrn, err := (*knirv.NFT)GetNRNByNFTID(1235, "NFT")
//	if err != nil {
//		return nil, fmt.Errorf("failed to resolve nrn: %w", err)
//	}
//	return nrn, nil
//  default:
//	return nil, fmt.Errorf("invalid resource in uri: %s", resource)
//}

//}
