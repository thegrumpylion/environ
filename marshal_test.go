package environ

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	conf := &struct {
		Addr string
		Port int
	}{
		"someOtherHost.net",
		8090,
	}
	out, err := Marshal(conf)

	require := require.New(t)

	require.Nil(err)
	require.Equal([]string{
		"ADDR=someOtherHost.net",
		"PORT=8090",
	}, out)
}

func TestMarshalMap(t *testing.T) {
	conf := &struct {
		Addr string
		Port int
	}{
		"someOtherHost.net",
		8090,
	}
	out, err := MarshalMap(conf)

	require := require.New(t)

	require.Nil(err)
	require.Equal(out["ADDR"], "someOtherHost.net")
	require.Equal(out["PORT"], "8090")
}

func TestMarshalPfx(t *testing.T) {
	conf := &struct {
		Addr string
		Port int
	}{
		"someOtherHost.net",
		8090,
	}
	out, err := MarshalPfx(conf, "__PFX__")

	require := require.New(t)

	require.Nil(err)
	require.Equal([]string{
		"__PFX__ADDR=someOtherHost.net",
		"__PFX__PORT=8090",
	}, out)
}
