package torrent

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	FileName = "debain-iso.torrent"
)

func TestTorrentFile(t *testing.T) {
	file, err := os.Open(FileName)
	assert.Nil(t, err)

	res, err := ParseTorrentFile(file)
	assert.Nil(t, err)
	fmt.Printf("res: %v\n", res)

}
