package user

import (
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stashapp/stashdb/pkg/email"
	"github.com/stashapp/stashdb/pkg/manager/config"
	"github.com/stashapp/stashdb/pkg/models"
)

var ErrInvalidActivationKey = errors.New("invalid activation key")

// NewUser registers a new user. It returns the activation key only if
// email verification is not required, otherwise it returns nil.
func NewUser(tx *sqlx.Tx, em *email.Manager, email, inviteKey string) (*string, error) {
	if err := ClearExpiredActivations(tx); err != nil {
		return nil, err
	}

	// ensure user or pending activation with email does not already exist
	uqb := models.NewUserQueryBuilder(tx)
	aqb := models.NewPendingActivationQueryBuilder(tx)
	iqb := models.NewInviteCodeQueryBuilder(tx)

	if err := validateUserEmail(email); err != nil {
		return nil, err
	}

	if err := validateExistingEmail(&uqb, &aqb, email); err != nil {
		return nil, err
	}

	// if existing activation exists with the same email, then re-create it
	a, err := aqb.FindByEmail(email, models.PendingActivationTypeNewUser)
	if err != nil {
		return nil, err
	}

	if a != nil {
		if err := aqb.Destroy(a.ID); err != nil {
			return nil, err
		}
	}

	inviteID, err := validateInviteKey(&iqb, &aqb, inviteKey)
	if err != nil {
		return nil, err
	}

	// generate an activation key and email
	key, err := generateActivationKey(&aqb, email, inviteID)
	if err != nil {
		return nil, err
	}

	// if activation is not required, then return the activation key
	if !config.GetRequireActivation() {
		return &key, nil
	}

	if err := sendNewUserEmail(em, email, key); err != nil {
		return nil, err
	}

	return nil, nil
}

func validateExistingEmail(f models.UserFinder, aqb models.PendingActivationFinder, email string) error {
	u, err := f.FindByEmail(email)
	if err != nil {
		return err
	}

	if u != nil {
		return errors.New("email already in use")
	}

	return nil
}

func validateInviteKey(iqb models.InviteKeyFinder, aqb models.PendingActivationFinder, inviteKey string) (uuid.NullUUID, error) {
	var ret uuid.NullUUID
	if config.GetRequireInvite() {
		if inviteKey == "" {
			return ret, errors.New("invite key required")
		}

		var err error
		ret.UUID, _ = uuid.FromString(inviteKey)
		ret.Valid = true

		key, err := iqb.Find(ret.UUID)
		if err != nil {
			return ret, err
		}

		if key == nil {
			return ret, errors.New("invalid invite key")
		}

		// ensure key isn't already used
		a, err := aqb.FindByInviteKey(inviteKey, models.PendingActivationTypeNewUser)
		if err != nil {
			return ret, err
		}

		if a != nil {
			return ret, errors.New("key already used")
		}
	}

	return ret, nil
}

func generateActivationKey(aqb models.PendingActivationCreator, email string, inviteKey uuid.NullUUID) (string, error) {
	UUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	currentTime := time.Now()

	activation := models.PendingActivation{
		ID:        UUID,
		Email:     email,
		InviteKey: inviteKey,
		Time: models.SQLiteTimestamp{
			Timestamp: currentTime,
		},
		Type: models.PendingActivationTypeNewUser,
	}

	obj, err := aqb.Create(activation)
	if err != nil {
		return "", err
	}

	return obj.ID.String(), nil
}

func ClearExpiredActivations(tx *sqlx.Tx) error {
	expireTime := config.GetActivationExpireTime()

	aqb := models.NewPendingActivationQueryBuilder(tx)
	return aqb.DestroyExpired(expireTime)
}

func sendNewUserEmail(em *email.Manager, email, activationKey string) error {
	subject := "Subject: Activate stash-box account"

	link := config.GetHostURL() + "/activate?email=" + email + "&key=" + activationKey
	body := "Please click the following link to activate your account: " + link

	return em.Send(email, subject, body)
}

func ActivateNewUser(tx *sqlx.Tx, name, email, activationKey, password string) (*models.User, error) {
	if err := ClearExpiredActivations(tx); err != nil {
		return nil, err
	}

	id, _ := uuid.FromString(activationKey)

	uqb := models.NewUserQueryBuilder(tx)
	aqb := models.NewPendingActivationQueryBuilder(tx)
	iqb := models.NewInviteCodeQueryBuilder(tx)

	a, err := aqb.Find(id)
	if err != nil {
		return nil, err
	}

	if a == nil || a.Email != email || a.Type != models.PendingActivationTypeNewUser {
		return nil, ErrInvalidActivationKey
	}

	// check expiry

	i, err := iqb.Find(a.InviteKey.UUID)
	if err != nil {
		return nil, err
	}

	if i == nil {
		return nil, errors.New("cannot find invite key")
	}

	invitedBy := i.GeneratedBy.String()

	createInput := models.UserCreateInput{
		Name:        name,
		Email:       email,
		Password:    password,
		InvitedByID: &invitedBy,
		Roles:       getDefaultUserRoles(),
	}

	if err := ValidateCreate(createInput); err != nil {
		return nil, err
	}

	// ensure user name does not already exist
	u, err := uqb.FindByName(name)
	if err != nil {
		return nil, err
	}

	if u != nil {
		return nil, errors.New("username already used")
	}

	ret, err := Create(tx, createInput)
	if err != nil {
		return nil, err
	}

	// delete the activation
	if err := aqb.Destroy(id); err != nil {
		return nil, err
	}

	// delete the invite key
	if err := iqb.Destroy(a.InviteKey.UUID); err != nil {
		return nil, err
	}

	return ret, nil
}

// ResetPassword generates an email to reset a users password.
func ResetPassword(tx *sqlx.Tx, em *email.Manager, email string) error {
	uqb := models.NewUserQueryBuilder(tx)
	aqb := models.NewPendingActivationQueryBuilder(tx)

	// ensure user exists
	u, err := uqb.FindByEmail(email)
	if err != nil {
		return err
	}

	if u == nil {
		// return silently
		return nil
	}

	// if existing activation exists with the same email, then re-create it
	a, err := aqb.FindByEmail(email, models.PendingActivationTypeResetPassword)
	if err != nil {
		return err
	}

	if a != nil {
		if err := aqb.Destroy(a.ID); err != nil {
			return err
		}
	}

	// generate an activation key and email
	key, err := generateResetPasswordActivationKey(&aqb, email)
	if err != nil {
		return err
	}

	if err := sendResetPasswordEmail(em, email, key); err != nil {
		return err
	}

	return nil
}

func generateResetPasswordActivationKey(aqb models.PendingActivationCreator, email string) (string, error) {
	UUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	currentTime := time.Now()

	activation := models.PendingActivation{
		ID:    UUID,
		Email: email,
		Time: models.SQLiteTimestamp{
			Timestamp: currentTime,
		},
		Type: models.PendingActivationTypeResetPassword,
	}

	obj, err := aqb.Create(activation)
	if err != nil {
		return "", err
	}

	return obj.ID.String(), nil
}

func sendResetPasswordEmail(em *email.Manager, email, activationKey string) error {
	subject := "Subject: Reset stash-box password"

	link := config.GetHostURL() + "/resetPassword?email=" + email + "&key=" + activationKey
	body := "Please click the following link to set your account password: " + link

	return em.Send(email, subject, body)
}

func ActivateResetPassword(tx *sqlx.Tx, activationKey string, newPassword string) error {
	if err := ClearExpiredActivations(tx); err != nil {
		return err
	}

	id, _ := uuid.FromString(activationKey)

	uqb := models.NewUserQueryBuilder(tx)
	aqb := models.NewPendingActivationQueryBuilder(tx)

	a, err := aqb.Find(id)
	if err != nil {
		return err
	}

	if a == nil || a.Type != models.PendingActivationTypeResetPassword {
		return ErrInvalidActivationKey
	}

	user, err := uqb.FindByEmail(a.Email)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user does not exist")
	}

	err = validateUserPassword(user.Name, user.Email, newPassword)
	if err != nil {
		return err
	}

	err = user.SetPasswordHash(newPassword)
	if err != nil {
		return err
	}
	user.UpdatedAt = models.SQLiteTimestamp{Timestamp: time.Now()}

	user, err = uqb.Update(*user)
	if err != nil {
		return err
	}

	// delete the activation
	if err := aqb.Destroy(id); err != nil {
		return err
	}

	return nil
}
