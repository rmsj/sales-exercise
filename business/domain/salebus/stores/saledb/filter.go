package saledb

import (
	"bytes"
	"strings"

	"github.com/rmsj/service/business/domain/salebus"
)

func (s *Store) applyFilter(filter salebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
