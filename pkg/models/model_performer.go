package models

import (
	"database/sql"
	"github.com/gofrs/uuid"

	"github.com/stashapp/stashdb/pkg/database"
)

const (
	performerTable   = "performers"
	performerJoinKey = "performer_id"
)

var (
	performerDBTable = database.NewTable(performerTable, func() interface{} {
		return &Performer{}
	})

	performerAliasTable = database.NewTableJoin(performerTable, "performer_aliases", performerJoinKey, func() interface{} {
		return &PerformerAlias{}
	})

	performerUrlTable = database.NewTableJoin(performerTable, "performer_urls", performerJoinKey, func() interface{} {
		return &PerformerUrl{}
	})

	performerTattooTable = database.NewTableJoin(performerTable, "performer_tattoos", performerJoinKey, func() interface{} {
		return &PerformerBodyMod{}
	})

	performerPiercingTable = database.NewTableJoin(performerTable, "performer_piercings", performerJoinKey, func() interface{} {
		return &PerformerBodyMod{}
	})
)

type Performer struct {
	ID                uuid.UUID       `db:"id" json:"id"`
	Name              string          `db:"name" json:"name"`
	Disambiguation    sql.NullString  `db:"disambiguation" json:"disambiguation"`
	Gender            sql.NullString  `db:"gender" json:"gender"`
	Birthdate         SQLiteDate      `db:"birthdate" json:"birthdate"`
	BirthdateAccuracy sql.NullString  `db:"birthdate_accuracy" json:"birthdate_accuracy"`
	Ethnicity         sql.NullString  `db:"ethnicity" json:"ethnicity"`
	Country           sql.NullString  `db:"country" json:"country"`
	EyeColor          sql.NullString  `db:"eye_color" json:"eye_color"`
	HairColor         sql.NullString  `db:"hair_color" json:"hair_color"`
	Height            sql.NullInt64   `db:"height" json:"height"`
	CupSize           sql.NullString  `db:"cup_size" json:"cup_size"`
	BandSize          sql.NullInt64   `db:"band_size" json:"band_size"`
	WaistSize         sql.NullInt64   `db:"waist_size" json:"waist_size"`
	HipSize           sql.NullInt64   `db:"hip_size" json:"hip_size"`
	BreastType        sql.NullString  `db:"breast_type" json:"breast_type"`
	CareerStartYear   sql.NullInt64   `db:"career_start_year" json:"career_start_year"`
	CareerEndYear     sql.NullInt64   `db:"career_end_year" json:"career_end_year"`
	CreatedAt         SQLiteTimestamp `db:"created_at" json:"created_at"`
	UpdatedAt         SQLiteTimestamp `db:"updated_at" json:"updated_at"`
	Deleted           bool            `db:"deleted" json:"deleted"`
}

func (Performer) GetTable() database.Table {
	return performerDBTable
}

func (p Performer) GetID() uuid.UUID {
	return p.ID
}

type Performers []*Performer

func (p Performers) Each(fn func(interface{})) {
	for _, v := range p {
		fn(*v)
	}
}

func (p *Performers) Add(o interface{}) {
	*p = append(*p, o.(*Performer))
}

type PerformerAlias struct {
	PerformerID uuid.UUID `db:"performer_id" json:"performer_id"`
	Alias       string    `db:"alias" json:"alias"`
}

type PerformerAliases []*PerformerAlias

func (p PerformerAliases) Each(fn func(interface{})) {
	for _, v := range p {
		fn(*v)
	}
}

func (p *PerformerAliases) Add(o interface{}) {
	*p = append(*p, o.(*PerformerAlias))
}

func (p PerformerAliases) ToAliases() []string {
	var ret []string
	for _, v := range p {
		ret = append(ret, v.Alias)
	}

	return ret
}

func CreatePerformerAliases(performerId uuid.UUID, aliases []string) PerformerAliases {
	var ret PerformerAliases

	for _, alias := range aliases {
		ret = append(ret, &PerformerAlias{PerformerID: performerId, Alias: alias})
	}

	return ret
}

type PerformerUrl struct {
	PerformerID uuid.UUID `db:"performer_id" json:"performer_id"`
	URL         string    `db:"url" json:"url"`
	Type        string    `db:"type" json:"type"`
}

func (p *PerformerUrl) ToURL() URL {
	url := URL{
		URL:  p.URL,
		Type: p.Type,
	}
	return url
}

type PerformerUrls []*PerformerUrl

func (p PerformerUrls) Each(fn func(interface{})) {
	for _, v := range p {
		fn(*v)
	}
}

func (p *PerformerUrls) Add(o interface{}) {
	*p = append(*p, o.(*PerformerUrl))
}

func CreatePerformerUrls(performerId uuid.UUID, urls []*URLInput) PerformerUrls {
	var ret PerformerUrls

	for _, urlInput := range urls {
		ret = append(ret, &PerformerUrl{
			PerformerID: performerId,
			URL:         urlInput.URL,
			Type:        urlInput.Type,
		})
	}

	return ret
}

type PerformerBodyMod struct {
	PerformerID uuid.UUID      `db:"performer_id" json:"performer_id"`
	Location    string         `db:"location" json:"location"`
	Description sql.NullString `db:"description" json:"description"`
}

func (m PerformerBodyMod) ToBodyModification() BodyModification {
	ret := BodyModification{
		Location: m.Location,
	}
	if m.Description.Valid {
		ret.Description = &(m.Description.String)
	}

	return ret
}

type PerformerBodyMods []*PerformerBodyMod

func (p PerformerBodyMods) Each(fn func(interface{})) {
	for _, v := range p {
		fn(*v)
	}
}

func (p *PerformerBodyMods) Add(o interface{}) {
	*p = append(*p, o.(*PerformerBodyMod))
}

func CreatePerformerBodyMods(performerId uuid.UUID, urls []*BodyModificationInput) PerformerBodyMods {
	var ret PerformerBodyMods

	for _, bmInput := range urls {
		description := sql.NullString{}

		if bmInput.Description != nil {
			description.String = *bmInput.Description
			description.Valid = true
		}
		ret = append(ret, &PerformerBodyMod{
			PerformerID: performerId,
			Location:    bmInput.Location,
			Description: description,
		})
	}

	return ret
}

func (p *Performer) IsEditTarget() {
}

func (p *Performer) setBirthdate(fuzzyDate FuzzyDateInput) {
	p.Birthdate = SQLiteDate{String: fuzzyDate.Date, Valid: fuzzyDate.Date != ""}
	p.BirthdateAccuracy = sql.NullString{String: fuzzyDate.Accuracy.String(), Valid: fuzzyDate.Date != ""}
}

func (p Performer) ResolveBirthdate() FuzzyDate {
	ret := FuzzyDate{}

	if p.Birthdate.Valid {
		ret.Date = p.Birthdate.String
	}
	if p.BirthdateAccuracy.Valid {
		ret.Accuracy = DateAccuracyEnum(p.BirthdateAccuracy.String)
		if !ret.Accuracy.IsValid() {
			ret.Accuracy = ""
		}
	}

	return ret
}

func (p *Performer) setMeasurements(measurements MeasurementsInput) {
	if measurements.CupSize != nil {
		p.CupSize = sql.NullString{String: *measurements.CupSize, Valid: *measurements.CupSize != ""}
	}
	if measurements.BandSize != nil {
		p.BandSize = sql.NullInt64{Int64: int64(*measurements.BandSize), Valid: *measurements.BandSize != 0}
	}
	if measurements.Hip != nil {
		p.HipSize = sql.NullInt64{Int64: int64(*measurements.Hip), Valid: *measurements.Hip != 0}
	}
	if measurements.Waist != nil {
		p.WaistSize = sql.NullInt64{Int64: int64(*measurements.Waist), Valid: *measurements.Waist != 0}
	}
}

func (p Performer) ResolveMeasurements() Measurements {
	ret := Measurements{}

	if p.CupSize.Valid {
		ret.CupSize = &p.CupSize.String
	}
	if p.BandSize.Valid {
		i := int(p.BandSize.Int64)
		ret.BandSize = &i
	}
	if p.HipSize.Valid {
		i := int(p.HipSize.Int64)
		ret.Hip = &i
	}
	if p.WaistSize.Valid {
		i := int(p.WaistSize.Int64)
		ret.Waist = &i
	}

	return ret
}

func (p *Performer) CopyFromCreateInput(input PerformerCreateInput) error {
	CopyFull(p, input)

	if input.Birthdate != nil {
		p.setBirthdate(*input.Birthdate)
	}

	if input.Measurements != nil {
		p.setMeasurements(*input.Measurements)
	}

	return nil
}

func (p *Performer) CopyFromUpdateInput(input PerformerUpdateInput) error {
	CopyFull(p, input)

	if input.Birthdate != nil {
		p.setBirthdate(*input.Birthdate)
	}

	if input.Measurements != nil {
		p.setMeasurements(*input.Measurements)
	}

	return nil
}

func CreatePerformerImages(performerID uuid.UUID, imageIds []string) PerformerImages {
	var imageJoins PerformerImages
	for _, iid := range imageIds {
		imageID := uuid.FromStringOrNil(iid)
		imageJoin := &PerformerImage{
			PerformerID: performerID,
			ImageID:     imageID,
		}
		imageJoins = append(imageJoins, imageJoin)
	}

	return imageJoins
}
