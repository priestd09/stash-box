package api

import (
	"context"
	"time"

	"github.com/stashapp/stashdb/pkg/dataloader"
	"github.com/stashapp/stashdb/pkg/models"
	"github.com/stashapp/stashdb/pkg/utils"
)

type performerResolver struct{ *Resolver }

func (r *performerResolver) ID(ctx context.Context, obj *models.Performer) (string, error) {
	return obj.ID.String(), nil
}

func (r *performerResolver) Disambiguation(ctx context.Context, obj *models.Performer) (*string, error) {
	return resolveNullString(obj.Disambiguation), nil
}

func (r *performerResolver) Aliases(ctx context.Context, obj *models.Performer) ([]string, error) {
	return dataloader.For(ctx).PerformerAliasesById.Load(obj.ID)
}

func (r *performerResolver) Gender(ctx context.Context, obj *models.Performer) (*models.GenderEnum, error) {
	var ret models.GenderEnum
	if !resolveEnum(obj.Gender, &ret) {
		return nil, nil
	}

	return &ret, nil
}

func (r *performerResolver) Urls(ctx context.Context, obj *models.Performer) ([]*models.URL, error) {
	return dataloader.For(ctx).PerformerUrlsById.Load(obj.ID)
}

func (r *performerResolver) Birthdate(ctx context.Context, obj *models.Performer) (*models.FuzzyDate, error) {
	ret := obj.ResolveBirthdate()
	return &ret, nil
}

func (r *performerResolver) Age(ctx context.Context, obj *models.Performer) (*int, error) {
	if !obj.Birthdate.Valid {
		return nil, nil
	}

	birthdate, err := utils.ParseDateStringAsTime(obj.Birthdate.String)
	if err != nil {
		return nil, nil
	}

	birthYear := birthdate.Year()
	now := time.Now()
	thisYear := now.Year()
	age := thisYear - birthYear
	if now.YearDay() < birthdate.YearDay() {
		age = age - 1
	}

	return &age, nil
}

func (r *performerResolver) Ethnicity(ctx context.Context, obj *models.Performer) (*models.EthnicityEnum, error) {
	var ret models.EthnicityEnum
	if !resolveEnum(obj.Ethnicity, &ret) {
		return nil, nil
	}

	return &ret, nil
}

func (r *performerResolver) Country(ctx context.Context, obj *models.Performer) (*string, error) {
	return resolveNullString(obj.Country), nil
}

func (r *performerResolver) EyeColor(ctx context.Context, obj *models.Performer) (*models.EyeColorEnum, error) {
	var ret models.EyeColorEnum
	if !resolveEnum(obj.EyeColor, &ret) {
		return nil, nil
	}

	return &ret, nil
}

func (r *performerResolver) HairColor(ctx context.Context, obj *models.Performer) (*models.HairColorEnum, error) {
	var ret models.HairColorEnum
	if !resolveEnum(obj.HairColor, &ret) {
		return nil, nil
	}

	return &ret, nil
}

func (r *performerResolver) Height(ctx context.Context, obj *models.Performer) (*int, error) {
	return resolveNullInt64(obj.Height)
}

func (r *performerResolver) Measurements(ctx context.Context, obj *models.Performer) (*models.Measurements, error) {
	ret := obj.ResolveMeasurements()
	return &ret, nil
}

func (r *performerResolver) BreastType(ctx context.Context, obj *models.Performer) (*models.BreastTypeEnum, error) {
	var ret models.BreastTypeEnum
	if !resolveEnum(obj.BreastType, &ret) {
		return nil, nil
	}

	return &ret, nil
}

func (r *performerResolver) CareerStartYear(ctx context.Context, obj *models.Performer) (*int, error) {
	return resolveNullInt64(obj.CareerStartYear)
}

func (r *performerResolver) CareerEndYear(ctx context.Context, obj *models.Performer) (*int, error) {
	return resolveNullInt64(obj.CareerEndYear)
}

func (r *performerResolver) Tattoos(ctx context.Context, obj *models.Performer) ([]*models.BodyModification, error) {
	return dataloader.For(ctx).PerformerTattoosById.Load(obj.ID)
}

func (r *performerResolver) Piercings(ctx context.Context, obj *models.Performer) ([]*models.BodyModification, error) {
	return dataloader.For(ctx).PerformerPiercingsById.Load(obj.ID)
}

func (r *performerResolver) Images(ctx context.Context, obj *models.Performer) ([]*models.Image, error) {
	imageIDs, err := dataloader.For(ctx).PerformerImageIDsById.Load(obj.ID)
	if err != nil {
		return nil, err
	}
	images, errors := dataloader.For(ctx).ImageById.LoadAll(imageIDs)
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}
	return images, nil
}
