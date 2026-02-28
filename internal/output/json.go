package output

import (
	"encoding/json"
	"os"

	"hunterX/internal/models"
)

func SaveJSON(filename string, data []models.Recruiter) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
