package environ

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKitchenSink(t *testing.T) {

	type addrType int
	const addrHome = 0
	const addrWork = 1

	out := &struct {
		Name   string
		Age    uint8
		List   []int
		Fruits []string
		Addr   []struct {
			Type   addrType
			Street string
		}
		RecPtr *struct {
			A string
			B int
			S []*string
		}
		PortName map[uint16]string
		ArrList  [][]string
	}{}

	envMap := map[string]string{
		"NAME":          "Lufy",
		"AGE":           "66",
		"LIST":          "34,65,234,653,3",
		"FRUITS":        "banana,apple,lemon,whatever",
		"ADDR.0.TYPE":   "1",
		"ADDR.0.STREET": "Somewhere str. over the rainbow",
		"ADDR.1.TYPE":   "0",
		"ADDR.1.STREET": "Somewhereelse str. over.",
		"RECPTR.A":      "ValA",
		"RECPTR.B":      "-1002",
		"RECPTR.S":      "this,is,a,ptr,list",
		"PORTNAME.80":   "http",
		"PORTNAME.443":  "https",
		"PORTNAME.25":   "smtp",
		"ARRLIST.0":     "list,of,strings",
		"ARRLIST.1":     "once,upon,\"a time, comma\",omg",
		"ARRLIST.2":     "red,green,blue",
	}

	env := []string{}
	for k, v := range envMap {
		env = append(env, k+"="+v)
	}

	err := UnmarshalEnv(out, env, "")
	if err != nil {
		t.Fatalf("err: %v\n", err)
	}

	require := require.New(t)

	require.Equal("Lufy", out.Name)
	require.EqualValues(66, out.Age)
	require.Equal([]int{34, 65, 234, 653, 3}, out.List, 0)
	require.Equal([]string{"banana", "apple", "lemon", "whatever"}, out.Fruits, 0)
	require.EqualValues(addrWork, out.Addr[0].Type)
	require.Equal("Somewhere str. over the rainbow", out.Addr[0].Street)
	require.EqualValues(addrHome, out.Addr[1].Type)
	require.Equal("Somewhereelse str. over.", out.Addr[1].Street)
	require.Equal("ValA", out.RecPtr.A)
	require.Equal(-1002, out.RecPtr.B)

	vals := []string{"this", "is", "a", "ptr", "list"}
	strp := []*string{}
	for _, v := range vals {
		ptr := new(string)
		*ptr = v
		strp = append(strp, ptr)
	}
	require.Equal(strp, out.RecPtr.S)

	m := map[uint16]string{
		80:  "http",
		443: "https",
		25:  "smtp",
	}
	require.Equal(m, out.PortName)

	arr := [][]string{}
	arr0 := []string{"list", "of", "strings"}
	arr = append(arr, arr0)
	arr1 := []string{"once", "upon", "a time, comma", "omg"}
	arr = append(arr, arr1)
	arr2 := []string{"red", "green", "blue"}
	arr = append(arr, arr2)
	require.Equal(arr, out.ArrList)

	// now, reverse it
	ret, err := MarshalEnv(out, "")
	require.Nil(err)
	require.ElementsMatch(env, ret)
}
