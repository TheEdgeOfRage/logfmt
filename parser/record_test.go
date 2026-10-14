package parser

import "testing"

func TestMatchesFilterEmptyValue(t *testing.T) {
	record := &Record{fields: map[string]string{"msg": "hello"}}
	filter := map[string]string{"status": ""}

	if record.MatchesFilter(filter) {
		t.Fatal("missing field matched empty-value filter")
	}

	record.fields["status"] = ""
	if !record.MatchesFilter(filter) {
		t.Fatal("present empty field did not match empty-value filter")
	}
}
