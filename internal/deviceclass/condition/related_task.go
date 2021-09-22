package condition

type RelatedTask byte

const (
	ClassifyDevice RelatedTask = iota + 1
	PropertyVendor
	PropertyModel
	PropertyModelSeries
	PropertyDefault
)
