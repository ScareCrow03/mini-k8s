package uid

import (
	"strings"

	"github.com/google/uuid"
)

func NewUid() string {
	uid := uuid.New()
	uidstr := strings.Split(uid.String(), "-")[0]
	return uidstr
}
