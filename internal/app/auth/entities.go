package auth

// Claims are the normalised fields extracted from a verified ID token.
type Claims struct {
	Sub      string `json:"sub"` // stable identity key
	Locale   string `json:"locale"`
	Username string `json:"username"`
	Nonce    string `json:"nonce"`
}
