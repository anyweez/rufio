package shared

import (
	"bufio"
	"os"
	"strconv"
	// "strings"
)

func LoadIds(filename string) ([]int, error) {
	ids := make([]int, 0)

	// Read the file
	fp, oerr := os.Open(filename)
	defer fp.Close()

	if oerr != nil {
		return ids, oerr
	}

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		summ_id, perr := strconv.Atoi(scanner.Text())

		if perr != nil {
			return ids, perr
		}

		ids = append(ids, summ_id)
	}

	return ids, nil
}
