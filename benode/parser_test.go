package benode

import (
	"bufio"
	"fmt"
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
	var res NodeType = _testCtx.MustScan(rd)
	assert.NotNil(t, res)
	assert.Nil(t, _testCtx.Err())
}

func TestString(t *testing.T) {

	buff := `4:1234`

	rd := bufio.NewReader(strings.NewReader(buff))
	res := _testCtx.ScanString(rd)
	assert.Nil(t, _testCtx.Err())
	assert.Equal(t, 4, len(*res.data))
	fmt.Printf("res.data: %v\n", *res.data)

}

func TestInt(t *testing.T) {

	buff := `i1234e`

	rd := bufio.NewReader(strings.NewReader(buff))
	res := _testCtx.ScanInt(rd)
	assert.Nil(t, _testCtx.Err())
	assert.Equal(t, int64(1234), *res.data)
	fmt.Printf("res.data: %v\n", *res.data)
}

func TestList(t *testing.T) {

	input := `li1234el4:abcdee`
	rd := bufio.NewReader(strings.NewReader(input))
	res := _testCtx.ScanList(rd)
	assert.Nil(t, _testCtx.Err())
	assert.Equal(t, 2, len(res.data))
	assert.IsType(t, &ListNode{}, res.data[1])
	r := (res.data[1]).(*ListNode)
	assert.IsType(t, &StringNode{}, r.data[0])
	s := (r.data[0]).(*StringNode)
	assert.Equal(t, "abcd", *s.data)
	assert.IsType(t, &IntNode{}, res.data[0])
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
