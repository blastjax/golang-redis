package github

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayUsersTable(t *testing.T) {
	// Capture stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = stdout })

	err = DisplayUsersTable([]User{
		{ID: 1, Login: "mojombo", Type: "User", URL: "https://api.github.com/users/mojombo"},
		{ID: 2, Login: "defunkt", Type: "User", URL: "https://api.github.com/users/defunkt"},
	})
	require.NoError(t, err)
	w.Close()

	out, err := io.ReadAll(r)
	require.NoError(t, err)

	// Every header column has values, columns line up, and the summary comes after the rows
	assert.Equal(t, "ID  LOGIN    TYPE  URL\n"+
		"1   mojombo  User  https://api.github.com/users/mojombo\n"+
		"2   defunkt  User  https://api.github.com/users/defunkt\n"+
		"\n"+
		"Total users fetched: 2\n"+
		"Next since value: 2\n", string(out))
}
