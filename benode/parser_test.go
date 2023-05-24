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
	var res Benode = _testCtx.Scan(rd)
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

func TestMarshal(t *testing.T) {
	{
		str := `123456`
		res, err := Marshal(str)
		assert.Nil(t, err)
		r, ok := res.(*StringNode)
		assert.True(t, ok)
		assert.Equal(t, str, *r.data)
	}
	{
		i := 12345
		res, err := Marshal(i)
		assert.Nil(t, err)
		r, ok := res.(*IntNode)
		assert.True(t, ok)
		assert.Equal(t, int64(i), *r.data)
	}
	{
		l := []string{"aaa", "bbb", "ccc"}
		res, err := Marshal(l)
		assert.Nil(t, err)
		r, ok := res.(*ListNode)
		assert.True(t, ok)
		assert.Equal(t, len(l), len(r.data))
	}

	{
		d1 := map[string]int{
			"a": 123,
			"b": 456,
		}
		res, err := Marshal(d1)
		assert.Nil(t, err)
		r, ok := res.(*DictNode)
		assert.True(t, ok)
		assert.Equal(t, len(d1), len(r.data))
	}
	{
		type T1 struct {
			A int      `benode:"aaa"`
			B []string `benode:"bbb"`
		}
		d2 := &T1{
			A: 4321,
			B: []string{"a", "b", "c"},
		}
		res, err := Marshal(d2)
		assert.Nil(t, err)
		r, ok := res.(*DictNode)
		assert.True(t, ok)
		assert.Equal(t, 2, len(r.data))
	}

}

func TestDict(t *testing.T) {
	{
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
	{
		input := `d1:ii1234e1:q2:qq1:ali1ei2ei3eee`
		type test struct {
			A int     `benode:"i"`
			B string  `benode:"q"`
			C []int64 `benode:"a"`
		}
		rd := bufio.NewReader(strings.NewReader(input))
		res := _testCtx.ScanDict(rd)
		assert.Nil(t, _testCtx.Err())
		assert.Equal(t, 3, len(res.data))
		var out test
		err := res.Decode(&out)
		assert.Nil(t, err)
		assert.Equal(t, 1234, out.A)
		assert.Equal(t, "qq", out.B)
		assert.Equal(t, []int64{1, 2, 3}, out.C)
	}
}
