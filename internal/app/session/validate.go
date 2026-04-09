package session

func ValidateKey(key string) error {
	// *2 since hex encoded string
	if len(key) != SessionIDBytesLen*2 {
		return ErrInvalidSession
	}
	return nil
}
