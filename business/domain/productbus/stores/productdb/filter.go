package productdb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rmsj/service/business/domain/productbus"
)

func (s *Store) applyFilter(filter productbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Price != nil {
		data["price"] = filter.Price
		wc = append(wc, "price = :price")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
