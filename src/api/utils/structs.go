package utils

type Variant struct {
	chrom   int
	pos     int
	id      string
	ref     []string
	alt     []string
	format  string
	qual    int
	filter  string
	info    []Info
	samples []Sample
	fileId  string
}

type Info struct {
	id    string
	value string
}

type Sample struct {
	sampleId  string
	variation string
}
