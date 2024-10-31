package encryption

import (
	"encoding/base64"
	"fmt"
)

func Base64Decode(encodedData string) (string, error){
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("error decoding Base64 encoded data %v", err)
	}
	return string(decodedData), nil
}

func Base64Encode(data string) string {
	encodedDataWithoutPadding := base64.RawStdEncoding.EncodeToString([]byte(data))
	return encodedDataWithoutPadding
}