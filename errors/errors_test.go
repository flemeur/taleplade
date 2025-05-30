package errors_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flemeur/taleplade/errors"
)

func TestE(t *testing.T) {
	t.Parallel()

	cases := []struct {
		want       *errors.Error
		wantErrStr string
		name       string
		args       []any
	}{
		{
			name:       "Kind and message provided",
			args:       []any{errors.Internal, "This is a message for a human"},
			want:       &errors.Error{Kind: errors.Internal, Message: "This is a message for a human"},
			wantErrStr: "<internal error> This is a message for a human",
		},
		{
			name: "Op and error provided",
			args: []any{errors.Op("test.Test"), errors.New("this is a nested error")},
			want: &errors.Error{
				Op:  errors.Op("test.Test"),
				Err: errors.New("this is a nested error"),
			},
			wantErrStr: "test.Test: this is a nested error",
		},
		{
			name: "Op and *Error provided",
			args: []any{errors.Op("test.Test"), &errors.Error{Message: "This is a nested *Error"}},
			want: &errors.Error{
				Op:  errors.Op("test.Test"),
				Err: &errors.Error{Message: "This is a nested *Error"},
			},
			wantErrStr: "test.Test: This is a nested *Error",
		},
		{
			name: "Error provided",
			args: []any{fmt.Errorf("this is a nested error")},
			want: &errors.Error{
				Op:  errors.Op("errors_test.TestE.func1"),
				Err: fmt.Errorf("this is a nested error"),
			},
			wantErrStr: "errors_test.TestE.func1: this is a nested error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := errors.E(tc.args...)

			//nolint:errorlint
			got, ok := err.(*errors.Error)

			require.True(t, ok, "Expected underlying type of error to be *errors.Error")

			require.Equal(t, tc.want, got, "Expected errors to be equal")

			require.Equal(t, tc.wantErrStr, err.Error())
		})
	}
}

func TestEPanics(t *testing.T) {
	t.Parallel()

	require.PanicsWithValue(t, "call to errors.E with no arguments", func() { _ = errors.E() })
	require.PanicsWithValue(t, "invalid error Kind provided: 255 (unknown error)", func() {
		_ = errors.E(errors.Kind(math.MaxUint8))
	})
	require.PanicsWithValue(t, "invalid argument to errors.E of type int, value 12", func() { _ = errors.E(12) })
	require.PanicsWithValue(t, "message or kind must not be provided when an error is also provided", func() {
		_ = errors.E(errors.Internal, errors.New("this is a nested error"), "This is a message for a human")
	})
}

func TestEReturnsNilOnNilError(t *testing.T) {
	t.Parallel()

	var err error

	require.Nil(t, errors.E(nil))
	require.Nil(t, errors.E(err))
	require.Nil(t, errors.E(errors.Op("test.NilReturn"), nil))
	require.Nil(t, errors.E(errors.Op("test.NilReturn"), err))
}
