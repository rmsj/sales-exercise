package userdb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rmsj/service/business/domain/userbus"
)

func applyFilter(filter userbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["user_id"] = filter.ID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Email != nil {
		data["email"] = filter.Email.String()
		wc = append(wc, "email = :email")
	}

	if filter.StartCreatedDate != nil {
		data["start_created_at"] = filter.StartCreatedDate.UTC()
		wc = append(wc, "created_at >= :start_created_at")
	}

	if filter.EndCreatedDate != nil {
		data["end_created_at"] = filter.EndCreatedDate.UTC()
		wc = append(wc, "created_at <= :end_created_at")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
