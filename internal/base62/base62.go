package base62

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const base = uint64(len(alphabet))

func Encode(id uint64) string {
	if id == 0 {
		return string(alphabet[0])
	}
	var chars []byte
	for id > 0 {
		remainder := id % base
		chars = append(chars, alphabet[remainder])
		id = id / base
	}
	// Reverse
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}
