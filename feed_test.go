package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPlanet(t *testing.T) {
	pp, err := Load()

	require.NoError(t, err)
	require.Equal(t, "Charlie.Choe", pp.Author)
	require.Equal(t, "whitekid@gmail.com", pp.Email)
	require.NotEqual(t, 0, len(pp.Planets))

	require.Equal(t, "Go Planet", pp.Planets[0].Title)
	require.Equal(t, "golang.xml", pp.Planets[0].Output)
	for _, p := range pp.Planets {
		require.NotEqual(t, 0, len(p.Feeds))
	}
}
