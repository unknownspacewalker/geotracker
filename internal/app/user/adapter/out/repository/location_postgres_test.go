package repository_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.com/spacewalker/locations/internal/app/user/adapter/out/repository"
	"gitlab.com/spacewalker/locations/internal/app/user/core/domain"
	"gitlab.com/spacewalker/locations/internal/app/user/core/port"
	"testing"
	"time"
)

func (s *PostgresTestSuite) Test_PostgresQueries_SetLocation() {
	createUserArgs := []port.CreateUserArg{
		{
			Username: "user1",
		},
		{
			Username: "user2",
		},
	}
	users := s.seedUsers(createUserArgs)

	setLocationArgs := []port.SetLocationArg{
		{
			UserID:    users[0].ID,
			Latitude:  1.0,
			Longitude: 1.0,
		},
	}
	locations := s.seedLocations(setLocationArgs)

	testCases := []struct {
		name   string
		arg    port.SetLocationArg
		hasErr bool
		isErr  error
		asErr  error
		assert func(t *testing.T, user domain.Location, err error)
	}{
		{
			name: "OK_UserExist_LocationExist",
			arg: port.SetLocationArg{
				UserID:    users[0].ID,
				Latitude:  locations[0].Latitude + 1.0,
				Longitude: locations[0].Longitude + 1.0,
			},
			hasErr: false,
			isErr:  nil,
			asErr:  nil,
			assert: func(t *testing.T, location domain.Location, err error) {
				require.Equal(t, users[0].ID, location.UserID)
				require.Equal(t, locations[0].Latitude+1.0, location.Latitude)
				require.Equal(t, locations[0].Longitude+1.0, location.Longitude)
				require.WithinDuration(t, locations[0].CreatedAt, location.CreatedAt, time.Second)
				require.WithinDuration(t, locations[0].UpdatedAt, location.UpdatedAt, time.Second)
			},
		},
		{
			name: "OK_UserExist_LocationDoesNotExist",
			arg: port.SetLocationArg{
				UserID:    users[1].ID,
				Latitude:  1.0,
				Longitude: 1.0,
			},
			hasErr: false,
			isErr:  nil,
			asErr:  nil,
			assert: func(t *testing.T, location domain.Location, err error) {
				require.Equal(t, users[1].ID, location.UserID)
				require.Equal(t, 1.0, location.Latitude)
				require.Equal(t, 1.0, location.Longitude)
				require.WithinDuration(t, time.Now(), location.CreatedAt, time.Second)
				require.WithinDuration(t, time.Now(), location.UpdatedAt, time.Second)
			},
		},
		{
			name: "ErrForeignKey_UserDoesNotExist_LocationDoesNotExist",
			arg: port.SetLocationArg{
				UserID:    0,
				Latitude:  1.0,
				Longitude: 1.0,
			},
			hasErr: true,
			isErr:  nil,
			asErr:  nil,
			assert: func(t *testing.T, location domain.Location, err error) {
				require.Empty(t, location)
			},
		},
		{
			name: "ErrCheck_LattitudeLessThenMin",
			arg: port.SetLocationArg{
				UserID:    users[0].ID,
				Latitude:  -181.00,
				Longitude: 1.0,
			},
			hasErr: true,
			isErr:  nil,
			asErr:  &port.InvalidLocationError{},
			assert: func(t *testing.T, location domain.Location, err error) {
				var invalidLocationError *port.InvalidLocationError
				require.ErrorAs(t, err, &invalidLocationError)
				require.Equal(t, port.InvalidLocationError{
					Violations: []port.InvalidLocationErrorViolation{
						{
							Subject: "latitude",
							Value:   -181.00,
						},
					},
				}, *invalidLocationError)
			},
		},
		{
			name: "ErrCheck_LattitudeGreaterThenMax",
			arg: port.SetLocationArg{
				UserID:    users[0].ID,
				Latitude:  181.00,
				Longitude: 1.0,
			},
			hasErr: true,
			isErr:  nil,
			asErr:  &port.InvalidLocationError{},
			assert: func(t *testing.T, location domain.Location, err error) {
				var invalidLocationError *port.InvalidLocationError
				require.ErrorAs(t, err, &invalidLocationError)
				require.Equal(t, port.InvalidLocationError{
					Violations: []port.InvalidLocationErrorViolation{
						{
							Subject: "latitude",
							Value:   181.00,
						},
					},
				}, *invalidLocationError)
			},
		},
		{
			name: "ErrCheck_LongitudeLessThenMin",
			arg: port.SetLocationArg{
				UserID:    users[0].ID,
				Latitude:  1.0,
				Longitude: -91.00,
			},
			hasErr: true,
			isErr:  nil,
			asErr:  &port.InvalidLocationError{},
			assert: func(t *testing.T, location domain.Location, err error) {
				var invalidLocationError *port.InvalidLocationError
				require.ErrorAs(t, err, &invalidLocationError)
				require.Equal(t, port.InvalidLocationError{
					Violations: []port.InvalidLocationErrorViolation{
						{
							Subject: "longitude",
							Value:   -91.00,
						},
					},
				}, *invalidLocationError)
			},
		},
		{
			name: "ErrCheck_LongitudeGreaterThenMax",
			arg: port.SetLocationArg{
				UserID:    users[0].ID,
				Latitude:  1.0,
				Longitude: 91.00,
			},
			hasErr: true,
			isErr:  nil,
			asErr:  &port.InvalidLocationError{},
			assert: func(t *testing.T, location domain.Location, err error) {
				var invalidLocationError *port.InvalidLocationError
				require.ErrorAs(t, err, &invalidLocationError)
				require.Equal(t, port.InvalidLocationError{
					Violations: []port.InvalidLocationErrorViolation{
						{
							Subject: "longitude",
							Value:   91.00,
						},
					},
				}, *invalidLocationError)
			},
		},
	}

	repo := repository.NewPostgresRepository(s.db)

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			location, err := repo.SetLocation(context.Background(), tc.arg)
			if !tc.hasErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.isErr != nil {
					require.ErrorIs(t, err, tc.isErr)
				}
				if tc.asErr != nil {
					require.ErrorAs(t, err, &tc.asErr)
				}
			}
			tc.assert(t, location, err)
		})
	}
}
