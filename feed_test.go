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

	for _, p := range pp.Planets {
		require.Equal(t, "Go Planet", p.Title)
		require.Equal(t, "golang.xml", p.Output)
		require.NotEqual(t, 0, len(p.Feeds))
	}
}
