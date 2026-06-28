package chassis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	require.Equal(t, 2, Add(1, 1))
}
