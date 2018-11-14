package provider

// RecordType is a type of a DNS record.
type RecordType string

// DNS record types.
const (
	RecordTypeUnknown RecordType = ""
	RecordTypeA       RecordType = "A"
)

var typesMap = map[string]RecordType{
	RecordTypeA.String(): RecordTypeA,
}

// String returns record value.
func (t RecordType) String() string {
	return string(t)
}

// ParseRecordType returns record type from string value.
func ParseRecordType(v string) RecordType {
	if t, ok := typesMap[v]; ok {
		return t
	}
	return RecordTypeUnknown
}

// Record is a single DNS record.
type Record struct {
	ID     string
	Domain string
	Type   RecordType
	Name   string
	Value  string
}
