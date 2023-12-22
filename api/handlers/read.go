package api

import (
	"context"

	"gorm.io/gorm"
)

func (s *ApiHandler) GetAllAcounts(ctx context.Context, db gorm.DB) ([]*Account, error) {
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("accounts").
		Rows()
	if err != nil {
		return []*Account{}, err
	}
	response := []*Account{}
	for rows.Next() {
		account := Account{}
		err := rows.Scan(
			&account.ID,
			&account.FirstName,
			&account.Email,
			&account.Password,
      &account.CommunityID,
		)
		if err != nil {
			return []*Account{}, err
		}
		response = append(response, &account)
	}
	return response, nil
}

func (s *ApiHandler) GetAllComms(ctx context.Context, db gorm.DB) ([]*Community, error){
  rows, err := s.db.WithContext(ctx).
    Select("*").
    Table("communities").
    Rows()
  if err != nil {
    return []*Community{}, err
  }
  response := []*Community{}
  for rows.Next(){
    comm := Community{}
    err := rows.Scan(
      &comm.ID,
      &comm.CommunityName,
      &comm.NoOfUsers,
    )
    if err != nil{
      return []*Community{}, err
    }
    response = append(response, &comm)
  }
  return response, nil
}
