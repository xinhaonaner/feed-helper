package xml2csv

import (
	"encoding/csv"
	"encoding/xml"
	"io"
	"strings"
)

// Convert reads XML from src and writes CSV to dst. It uses streaming:
// each "row" element (e.g. <item>) is parsed and written as one CSV row.
// Column names are inferred from the first row's direct child element names.
func Convert(src io.Reader, dst io.Writer, opts ...Option) error {
	cfg := applyOptions(opts)
	dec := xml.NewDecoder(src)
	cw := csv.NewWriter(dst)
	defer cw.Flush()

	var columnOrder []string
	seenColumns := make(map[string]bool)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != cfg.RowTag {
			continue
		}

		row, order, err := readRow(dec, seenColumns, columnOrder)
		if err != nil {
			return err
		}
		if len(row) == 0 && len(order) == 0 {
			continue
		}

		if columnOrder == nil {
			columnOrder = order
			if err := cw.Write(columnOrder); err != nil {
				return err
			}
		}

		record := make([]string, len(columnOrder))
		for i, col := range columnOrder {
			record[i] = row[col]
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// readRow consumes one row element (from the current StartElement to its EndElement)
// and returns a map of column name -> value and the order of columns (for first row).
func readRow(dec *xml.Decoder, seenColumns map[string]bool, columnOrder []string) (map[string]string, []string, error) {
	row := make(map[string]string)
	var order []string
	depth := 1
	var currentCol string
	var currentText strings.Builder

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 2 {
				currentCol = t.Name.Local
				currentText.Reset()
			}
		case xml.EndElement:
			if depth == 2 {
				val := strings.TrimSpace(currentText.String())
				row[currentCol] = val
				if !seenColumns[currentCol] {
					seenColumns[currentCol] = true
					order = append(order, currentCol)
				}
				currentCol = ""
			}
			depth--
			if depth == 0 {
				return row, order, nil
			}
		case xml.CharData:
			if depth == 2 {
				currentText.Write(t)
			}
		}
	}
}
