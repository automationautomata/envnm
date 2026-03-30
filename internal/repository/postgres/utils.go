package postgres

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
