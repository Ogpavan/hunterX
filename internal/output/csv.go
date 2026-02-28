package output

import (
	"encoding/csv"
	"os"

	"hunterX/internal/models"
)

func SaveCSV(filename string, data []models.Recruiter) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{
		"Source",
		"CompanyName",
		"RecruiterName",
		"JobTitle",
		"Location",
		"Phone",
		"Email",
		"Website",
	})

	for _, r := range data {
		_ = w.Write([]string{
			r.Source,
			r.CompanyName,
			r.RecruiterName,
			r.JobTitle,
			r.Location,
			r.Phone,
			r.Email,
			r.Website,
		})
	}

	return nil
}
