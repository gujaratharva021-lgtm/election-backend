// repository/constituency_repository.go
package repository

import (
	"election_backend/models"

	"gorm.io/gorm"
)

type ConstituencyRepository struct {
	db *gorm.DB
}

func NewConstituencyRepository(db *gorm.DB) *ConstituencyRepository {
	return &ConstituencyRepository{db: db}
}

func (r *ConstituencyRepository) GetAll(state string) ([]models.Constituency, error) {
	var constituencies []models.Constituency
	query := r.db
	if state != "" {
		query = query.Where("state = ?", state)
	}
	err := query.Find(&constituencies).Error
	return constituencies, err
}

func (r *ConstituencyRepository) GetByID(id uint) (*models.Constituency, error) {
	var c models.Constituency
	err := r.db.First(&c, id).Error
	return &c, err
}

func (r *ConstituencyRepository) Save(c *models.Constituency) error {
	return r.db.Save(c).Error
}

func (r *ConstituencyRepository) SaveAll(cs []models.Constituency) error {
	return r.db.CreateInBatches(cs, 50).Error
}
