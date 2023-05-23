package benode

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	_testCtx = &NodeContextImpl{}
)

func TestDecode(t *testing.T) {
	buff := `d4:1234i5678eli8909e1:aedi2345e3:abcee`

	rd := bufio.NewReader(strings.NewReader(buff))
	var res Benode = _testCtx.MustScan(rd)
	assert.NotNil(t, res)
	assert.Nil(t, _testCtx.Err())
}

func TestString(t *testing.T) {

	buff := `4:1234`

	rd := bufio.NewReader(strings.NewReader(buff))
	res := _testCtx.ScanString(rd)
	assert.Nil(t, _testCtx.Err())
	assert.Equal(t, 4, len(*res.data))
	{
		var out string
		err := res.Decode(&out)
		assert.Nil(t, err)
		assert.Equal(t, "1234", out)

	}
	{
		var out1 float64
		err := res.Decode(&out1)
		assert.Nil(t, err)
		assert.Equal(t, float64(1234), out1)
	}

}

func TestInt(t *testing.T) {

	buff := `i1234e`

	rd := bufio.NewReader(strings.NewReader(buff))
	res := _testCtx.ScanInt(rd)
	assert.Nil(t, _testCtx.Err())

	var out int64
	err := res.Decode(&out)
	assert.Nil(t, err)
	assert.Equal(t, int64(1234), out)
}

func TestList(t *testing.T) {

	input := `li1234e4:abcde`
	rd := bufio.NewReader(strings.NewReader(input))
	res := _testCtx.ScanList(rd)
	assert.Nil(t, _testCtx.Err())
	assert.Equal(t, 2, len(res.data))

	{
		var out []string
		err := res.Decode(&out)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(out))

	}
}
func TestList1(t *testing.T) {

	input := `li1234el4:abcdee`
	rd := bufio.NewReader(strings.NewReader(input))
	res := _testCtx.ScanList(rd)
	assert.Nil(t, _testCtx.Err())
	assert.Equal(t, 2, len(res.data))

	{
		var out []any
		err := res.Decode(&out)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(out))
		assert.IsType(t, []any{}, out[1])

	}
}

func TestDict(t *testing.T) {

	input := `ddi1234e2:qqel4:abcdee`
	rd := bufio.NewReader(strings.NewReader(input))
	res := _testCtx.ScanDict(rd)
	assert.Nil(t, _testCtx.Err())
	assert.Equal(t, 1, len(res.data))
	for k, v := range res.data {
		assert.IsType(t, &DictNode{}, k)
		assert.IsType(t, &ListNode{}, v)
	}
}
