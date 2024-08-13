package guid

import (
	"regexp"
	"strings"
	"testing"
)

func TestS(t *testing.T) {
	// Test with no data
	id1 := S()
	if len(id1) != 32 {
		t.Errorf("Expected length 32, got %d", len(id1))
	}
	if !isValidGUID(id1) {
		t.Errorf("Invalid GUID format: %s", id1)
	}

	// Test with one data item
	data1 := []byte("test data")
	id2 := S(data1)
	if len(id2) != 32 {
		t.Errorf("Expected length 32, got %d", len(id2))
	}
	if !isValidGUID(id2) {
		t.Errorf("Invalid GUID format: %s", id2)
	}

	// Test with two data items
	data2 := []byte("more test data")
	id3 := S(data1, data2)
	if len(id3) != 32 {
		t.Errorf("Expected length 32, got %d", len(id3))
	}
	if !isValidGUID(id3) {
		t.Errorf("Invalid GUID format: %s", id3)
	}

	// Test uniqueness
	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Errorf("GUIDs are not unique")
	}
}

func TestGetSequence(t *testing.T) {
	seq1 := getSequence()
	seq2 := getSequence()

	if len(seq1) != 3 || len(seq2) != 3 {
		t.Errorf("Sequence should be 3 bytes long")
	}

	if string(seq1) == string(seq2) {
		t.Errorf("Consecutive sequences should not be equal")
	}
}

func TestGetRandomStr(t *testing.T) {
	rand1 := getRandomStr(6)
	rand2 := getRandomStr(6)

	if len(rand1) != 6 || len(rand2) != 6 {
		t.Errorf("Random string should be 6 bytes long")
	}

	if string(rand1) == string(rand2) {
		t.Errorf("Random strings should not be equal")
	}

	for _, c := range string(rand1) + string(rand2) {
		if !strings.ContainsRune(randomStrBase, c) {
			t.Errorf("Invalid character in random string: %c", c)
		}
	}
}

func TestGetDataHashStr(t *testing.T) {
	data := []byte("test data")
	hash := getDataHashStr(data)

	if len(hash) != 7 {
		t.Errorf("Data hash should be 7 bytes long")
	}

	hash2 := getDataHashStr(data)
	if string(hash) != string(hash2) {
		t.Errorf("Data hash should be deterministic")
	}
}

func isValidGUID(guid string) bool {
	return regexp.MustCompile(`^[0-9a-z]{32}$`).MatchString(guid)
}
