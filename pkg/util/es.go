package util

import (
	"fmt"
	"regexp"
	"strconv"
)

// NodeIndex attempts to extract the node index from the string `name`.
// It expects names to be of the form something-0, something-1, something-2.
func NodeIndex(name string) (int, error) {
	if len(name) == 0 {
		return -1, fmt.Errorf("node name must be set")
	}

	reg, err := regexp.Compile(`(\d)+$`)

	if err != nil {
		return -1, fmt.Errorf("error constructing regexp: %s", err.Error())
	}

	indexStr := reg.FindString(name)

	if len(indexStr) == 0 {
		return -1, fmt.Errorf("could not find node index in '%s'", name)
	}

	i, err := strconv.Atoi(indexStr)

	if err != nil {
		return -1, fmt.Errorf("error parsing node index '%s': %s", indexStr, err.Error())
	}

	return i, nil
}
