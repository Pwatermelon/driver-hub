package hubid

import (
	"crypto/rand"
	"fmt"
	"strings"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// Generate возвращает публичный идентификатор вида DH-XXXXXXXX.
func Generate() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("DH-")
	for _, x := range b {
		sb.WriteByte(alphabet[int(x)%len(alphabet)])
	}
	return sb.String(), nil
}

func Normalize(id string) string {
	return strings.ToUpper(strings.TrimSpace(id))
}

func Validate(id string) error {
	id = Normalize(id)
	if !strings.HasPrefix(id, "DH-") || len(id) != 11 {
		return fmt.Errorf("некорректный Hub ID: ожидается формат DH-XXXXXXXX")
	}
	return nil
}
