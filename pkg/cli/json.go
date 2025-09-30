package cli

import (
	"encoding/json"
	"os"
)

func PrintJSON(res Result) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
}
