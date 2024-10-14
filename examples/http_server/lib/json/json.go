package json

import (
	"encoding/json"
	"fmt"
)

func Dumps(data interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}
	return jsonData
}

func Loads(jsonStr []byte) map[string]any {
	decodedData := map[string]any{"placeholder1": "", "placeholder2": 0}
	json.Unmarshal(jsonStr, &decodedData)
	delete(decodedData, "placeholder1")
	delete(decodedData, "placeholder2")
	return decodedData
}

