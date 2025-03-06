package io_cgp

import (
	"regexp"
)

var (
	re_output_delim = regexp.MustCompile(`[,\(\)]`)
)
